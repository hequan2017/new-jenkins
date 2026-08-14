package ops

import (
	"context"
	"encoding/json"
	"fmt"
	"io"

	"golang.org/x/crypto/ssh"
)

// 终端 <-> 前端的控制消息类型
const (
	TermMsgInput  = "input"  // 前端 -> 后端: 终端输入(按键)
	TermMsgResize = "resize" // 前端 -> 后端: 窗口尺寸变更
	TermMsgOutput = "output" // 后端 -> 前端: 终端输出
	TermMsgClose  = "close"  // 后端 -> 前端: 终端已关闭
	TermMsgError  = "error"  // 后端 -> 前端: 错误
)

// TerminalMessage 前后端约定的 WS 消息体
type TerminalMessage struct {
	Type string `json:"type"`
	Data string `json:"data"`
	Cols int    `json:"cols"`
	Rows int    `json:"rows"`
}

// TerminalService 交互式 SSH 终端: 把 SSH PTY 与一个双向消息通道桥接。
// 具体的传输层(WebSocket)由 utils/ws 提供, 调用方实现 TerminalConn 接口注入。
type TerminalService struct{}

// TerminalConn 抽象双向通道, 由 WebSocket handler 实现, 解耦传输与协议。
type TerminalConn interface {
	// Read 从前端读取一条消息(阻塞直到有消息或连接关闭)
	Read() (TerminalMessage, error)
	// Write 向前端发送一条消息
	Write(TerminalMessage) error
	// Close 关闭通道
	Close() error
	// Done 返回一个在连接关闭时解除阻塞的 channel
	Done() <-chan struct{}
}

// Serve 启动一个 SSH PTY 会话并桥接到 conn, 阻塞直到会话结束或连接关闭。
// cols/rows 为初始窗口尺寸。
func (s *TerminalService) Serve(ctx context.Context, assetID, credentialID uint, cols, rows int, conn TerminalConn) (retErr error) {
	client, err := buildSSHClient(ctx, assetID, credentialID)
	if err != nil {
		_ = conn.Write(TerminalMessage{Type: TermMsgError, Data: err.Error()})
		return err
	}
	defer func() {
		_ = client.Close()
		_ = conn.Write(TerminalMessage{Type: TermMsgClose, Data: "session ended"})
	}()

	sess, err := client.NewSession()
	if err != nil {
		_ = conn.Write(TerminalMessage{Type: TermMsgError, Data: "建立会话失败: " + err.Error()})
		return err
	}
	defer sess.Close()

	// 请求伪终端
	modes := ssh.TerminalModes{
		ssh.ECHO:          1,
		ssh.TTY_OP_ISPEED: 14400,
		ssh.TTY_OP_OSPEED: 14400,
	}
	if cols <= 0 {
		cols = 80
	}
	if rows <= 0 {
		rows = 24
	}
	if err := sess.RequestPty("xterm", rows, cols, modes); err != nil {
		_ = conn.Write(TerminalMessage{Type: TermMsgError, Data: "请求 PTY 失败: " + err.Error()})
		return err
	}

	stdin, err := sess.StdinPipe()
	if err != nil {
		return err
	}
	stdout, err := sess.StdoutPipe()
	if err != nil {
		return err
	}
	stderr, err := sess.StderrPipe()
	if err != nil {
		return err
	}

	if err := sess.Shell(); err != nil {
		_ = conn.Write(TerminalMessage{Type: TermMsgError, Data: "启动 shell 失败: " + err.Error()})
		return err
	}

	// 输出读取: stdout + stderr 汇聚到前端
	outDone := make(chan struct{})
	errDone := make(chan struct{})
	go func() {
		_, _ = io.Copy(pipeWriter{conn: conn}, stdout)
		close(outDone)
	}()
	go func() {
		_, _ = io.Copy(pipeWriter{conn: conn}, stderr)
		close(errDone)
	}()

	// 输入 + resize 读取: 从前端读取消息转发到 PTY stdin
	inputDone := make(chan struct{})
	go func() {
		defer close(inputDone)
		for {
			msg, err := conn.Read()
			if err != nil {
				return
			}
			switch msg.Type {
			case TermMsgInput:
				if _, werr := stdin.Write([]byte(msg.Data)); werr != nil {
					return
				}
			case TermMsgResize:
				w := msg.Cols
				h := msg.Rows
				if w <= 0 {
					w = cols
				}
				if h <= 0 {
					h = rows
				}
				_ = sess.WindowChange(h, w)
			}
		}
	}()

	// 等待会话结束 或 连接关闭 / context 取消
	waitErr := make(chan error, 1)
	go func() { waitErr <- sess.Wait() }()

	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-conn.Done():
		return nil
	case <-inputDone:
		return nil
	case err := <-waitErr:
		if err != nil {
			_ = conn.Write(TerminalMessage{Type: TermMsgError, Data: fmt.Sprintf("会话结束: %v", err)})
		}
		return err
	}
}

// pipeWriter 把 SSH 输出适配成向前端发送的 output 消息
type pipeWriter struct {
	conn TerminalConn
}

func (w pipeWriter) Write(p []byte) (int, error) {
	msg := TerminalMessage{Type: TermMsgOutput, Data: string(p)}
	// 忽略写入错误(连接可能已关闭), 不影响 SSH 侧读取计数
	_ = w.conn.Write(msg)
	return len(p), nil
}

// EncodeTerminalMsg 把终端消息编码为 JSON 字符串, 供 WS handler 作为 WS payload。
func EncodeTerminalMsg(m TerminalMessage) (string, error) {
	b, err := json.Marshal(m)
	if err != nil {
		return "", err
	}
	return string(b), nil
}

// DecodeTerminalMsg 把前端 JSON 字符串解码为终端消息。
func DecodeTerminalMsg(raw string) (TerminalMessage, error) {
	var m TerminalMessage
	err := json.Unmarshal([]byte(raw), &m)
	return m, err
}
