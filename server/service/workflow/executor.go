package workflow

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/url"
	"os"
	"os/exec"
	"runtime"
	"strings"
	"time"

	modelWorkflow "github.com/flipped-aurora/gin-vue-admin/server/model/workflow"
)

// StepResult 单个 Step 执行结果
type StepResult struct {
	ExitCode int    // shell 退出码; http 固定 0(成功)/1(失败)
	Output   string // 摘要(HTTP 响应片段或 shell stdout 概要), 真正日志走 logFn
	Err      error
}

// logFunc 日志采集回调: engine 实现并落库 + SSE 推送
// stream 取值 workflow.LogStreamStdout/Stderr/System。
type logFunc func(stream string, text string)

// StepExecutor 可插拔步骤执行器接口
// 本期提供 http / shell 两个实现; 后续可扩展(如 docker、远程 agent)。
// params 为本次构建的参数(键->值), 执行器在执行前对 config 做变量替换(${param.xxx})。
type StepExecutor interface {
	Execute(ctx context.Context, stepType string, config []byte, params map[string]string, log logFunc) StepResult
}

// defaultExecutor 按类型分派到具体执行器
type defaultExecutor struct {
	http  *httpExecutor
	shell *shellExecutor
}

func newDefaultExecutor() *defaultExecutor {
	return &defaultExecutor{http: &httpExecutor{}, shell: &shellExecutor{}}
}

func (d *defaultExecutor) Execute(ctx context.Context, stepType string, config []byte, params map[string]string, log logFunc) StepResult {
	// 变量替换: 把 config 里的 ${param.xxx} / $param.xxx 替换为实际参数值
	expanded := expandConfig(config, params)
	switch stepType {
	case modelWorkflow.StepTypeHTTP:
		return d.http.Execute(ctx, expanded, log)
	case modelWorkflow.StepTypeShell:
		return d.shell.Execute(ctx, expanded, log)
	default:
		return StepResult{ExitCode: 1, Err: fmt.Errorf("未知 step 类型: %s", stepType)}
	}
}

// expandConfig 对 step config 做 ${param.xxx} 变量替换。
// 用自定义映射函数支持 ${param.name} 和 $param.name 两种写法;
// 未知变量保留原样(不替换),避免误清空。
func expandConfig(config []byte, params map[string]string) []byte {
	if len(params) == 0 || len(config) == 0 {
		return config
	}
	expanded := os.Expand(string(config), func(key string) string {
		// 仅识别 param. 前缀的变量, 其它 $ 不替换(防止 shell 的 $VAR 被误替换)
		if strings.HasPrefix(key, "param.") {
			name := strings.TrimPrefix(key, "param.")
			if v, ok := params[name]; ok {
				return v
			}
		}
		// 未知变量: 还原成 ${key} 形式, 不替换
		return "${" + key + "}"
	})
	return []byte(expanded)
}

// ============================== HTTP 执行器 ==============================

type httpExecutorConfig struct {
	URL           string            `json:"url"`
	Method        string            `json:"method"`
	Headers       map[string]string `json:"headers"`
	Body          string            `json:"body"`
	TimeoutSec    int               `json:"timeoutSec"`
	AllowPrivate  bool              `json:"allowPrivate"`
	ExpectStatus  int               `json:"expectStatus"` // 期望状态码, 0 表示 2xx 即成功
}

type httpExecutor struct{}

func (h *httpExecutor) Execute(ctx context.Context, config []byte, log logFunc) StepResult {
	var cfg httpExecutorConfig
	if err := json.Unmarshal(config, &cfg); err != nil {
		log("system", "HTTP 步骤配置解析失败: "+err.Error())
		return StepResult{ExitCode: 1, Err: err}
	}
	if cfg.URL == "" {
		log("system", "HTTP 步骤缺少 url")
		return StepResult{ExitCode: 1, Err: fmt.Errorf("http 步骤缺少 url")}
	}
	u, err := url.Parse(cfg.URL)
	if err != nil || (u.Scheme != "http" && u.Scheme != "https") || u.Host == "" {
		log("system", "HTTP 步骤 url 非法: "+cfg.URL)
		return StepResult{ExitCode: 1, Err: fmt.Errorf("http url 非法")}
	}
	// SSRF 防护: 默认禁止内网/环回地址
	if !cfg.AllowPrivate && isPrivateOrLoopback(u.Hostname()) {
		log("system", "HTTP 步骤禁止访问内网/环回地址(可在配置中显式 allowPrivate=true 放行)")
		return StepResult{ExitCode: 1, Err: fmt.Errorf("禁止访问内网地址")}
	}
	method := strings.ToUpper(cfg.Method)
	if method == "" {
		method = "GET"
	}

	timeout := time.Duration(cfg.TimeoutSec) * time.Second
	if timeout <= 0 {
		timeout = 30 * time.Second
	}
	reqCtx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	var body io.Reader
	if cfg.Body != "" {
		body = strings.NewReader(cfg.Body)
	}
	req, err := http.NewRequestWithContext(reqCtx, method, cfg.URL, body)
	if err != nil {
		log("system", "构造 HTTP 请求失败: "+err.Error())
		return StepResult{ExitCode: 1, Err: err}
	}
	for k, v := range cfg.Headers {
		req.Header.Set(k, v)
	}

	log("stdout", fmt.Sprintf("%s %s", method, cfg.URL))
	client := &http.Client{Timeout: timeout}
	resp, err := client.Do(req)
	if err != nil {
		log("stderr", "HTTP 请求失败: "+err.Error())
		return StepResult{ExitCode: 1, Err: err}
	}
	defer resp.Body.Close()

	respBody, _ := io.ReadAll(io.LimitReader(resp.Body, 64*1024)) // 最多读 64KB 入日志
	if len(respBody) > 0 {
		log("stdout", "响应: "+string(respBody))
	}
	log("stdout", fmt.Sprintf("HTTP %d", resp.StatusCode))

	success := resp.StatusCode >= 200 && resp.StatusCode < 300
	if cfg.ExpectStatus > 0 {
		success = resp.StatusCode == cfg.ExpectStatus
	}
	if !success {
		return StepResult{ExitCode: 1, Err: fmt.Errorf("HTTP 状态码 %d", resp.StatusCode)}
	}
	return StepResult{ExitCode: 0}
}

// ============================== Shell 执行器 ==============================

type shellExecutorConfig struct {
	Command    string            `json:"command"`
	TimeoutSec int               `json:"timeoutSec"`
	Env        map[string]string `json:"env"`
}

type shellExecutor struct{}

func (s *shellExecutor) Execute(ctx context.Context, config []byte, log logFunc) StepResult {
	var cfg shellExecutorConfig
	if err := json.Unmarshal(config, &cfg); err != nil {
		log("system", "Shell 步骤配置解析失败: "+err.Error())
		return StepResult{ExitCode: 1, Err: err}
	}
	if strings.TrimSpace(cfg.Command) == "" {
		log("system", "Shell 步骤缺少 command")
		return StepResult{ExitCode: 1, Err: fmt.Errorf("shell 步骤缺少 command")}
	}

	timeout := time.Duration(cfg.TimeoutSec) * time.Second
	if timeout <= 0 {
		timeout = 10 * time.Minute
	}
	runCtx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	// 用系统 shell 执行(-c), 跨平台: Linux/macOS 用 sh, Windows 用 cmd。
	// 安全边界: command 由流水线编辑者填写, 视同 Jenkins 的 sh; 不做命令白名单,
	// 但通过子进程隔离 + 超时 + 输出采集控制其行为。
	log("stdout", "$ "+cfg.Command)
	// 统一用 "<shell> -c <command>" 形式; Windows 下优先 sh(git-bash 自带, 支持 unix 语法),
	// 找不到 sh 才退 cmd(此时流水线命令需用 cmd 语法)。
	shell, args := shellPath()
	cmd := exec.CommandContext(runCtx, shell, append(args, cfg.Command)...)
	if len(cfg.Env) > 0 {
		cmd.Env = append(cmd.Env, envSlice(cfg.Env)...)
	}
	cmd.Stdout = lineWriter{fn: func(line string) { log("stdout", line) }}
	cmd.Stderr = lineWriter{fn: func(line string) { log("stderr", line) }}

	if err := cmd.Run(); err != nil {
		if runCtx.Err() == context.DeadlineExceeded {
			log("stderr", "Shell 步骤执行超时")
			return StepResult{ExitCode: 124, Err: fmt.Errorf("shell 执行超时")}
		}
		// 非零退出码: 取 exit code
		if ee, ok := err.(*exec.ExitError); ok {
			log("stderr", fmt.Sprintf("退出码 %d", ee.ExitCode()))
			return StepResult{ExitCode: ee.ExitCode(), Err: err}
		}
		log("stderr", "Shell 执行失败: "+err.Error())
		return StepResult{ExitCode: 1, Err: err}
	}
	return StepResult{ExitCode: 0}
}

// shellPath 返回执行 shell 命令的解释器及其参数前缀。
// 优先 sh(跨平台一致语法, Linux 生产默认, Windows 开发环境 git-bash 自带);
// Windows 下若 PATH 无 sh 则回退 cmd(此时流水线命令需使用 cmd 语法)。
func shellPath() (string, []string) {
	if runtime.GOOS == "windows" {
		if _, err := exec.LookPath("sh"); err == nil {
			return "sh", []string{"-c"}
		}
		return "cmd", []string{"/c"}
	}
	return "sh", []string{"-c"}
}

// isPrivateOrLoopback 判断主机是否内网/环回地址(SSRF 防护)
// 解析失败或为 IP 字面量时按保守策略判定。
func isPrivateOrLoopback(host string) bool {
	if host == "" {
		return true
	}
	// 端口剥离
	if h, _, err := net.SplitHostPort(host); err == nil {
		host = h
	}
	ip := net.ParseIP(host)
	if ip == nil {
		// 域名: 尝试解析, 解析失败保守视为内网
		ips, err := net.LookupIP(host)
		if err != nil || len(ips) == 0 {
			return true
		}
		ip = ips[0]
	}
	return ip.IsLoopback() || ip.IsPrivate() || ip.IsLinkLocalUnicast() || ip.IsUnspecified()
}

func envSlice(m map[string]string) []string {
	out := make([]string, 0, len(m))
	for k, v := range m {
		out = append(out, k+"="+v)
	}
	return out
}

// lineWriter 把逐行输出转发给 logFunc; 简化实现: 直接整段转发
// (子进程输出量在引擎层不做行切分, 由 DB 存储结构保留; 后续可优化按行 SSE)
type lineWriter struct{ fn func(string) }

func (w lineWriter) Write(p []byte) (int, error) {
	w.fn(strings.TrimRight(string(p), "\n"))
	return len(p), nil
}
