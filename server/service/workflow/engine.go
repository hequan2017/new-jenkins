package workflow

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"
	"sync/atomic"
	"time"

	"github.com/flipped-aurora/gin-vue-admin/server/global"
	modelWorkflow "github.com/flipped-aurora/gin-vue-admin/server/model/workflow"
	"github.com/flipped-aurora/gin-vue-admin/server/utils/logger"
	"github.com/flipped-aurora/gin-vue-admin/server/utils/sse"
)

// Engine 流水线执行核心: Stage→Step 编排 + 状态机推进 + 日志采集 + SSE 推送
type Engine struct {
	executor StepExecutor

	// approval 等待表: buildID -> chan approvalSignal
	// 审批 gate 的 Stage 跑完前置 Step 后, Run 在此阻塞等待 NotifyApproval 唤醒。
	amu      sync.Mutex
	approval map[uint]chan approvalSignal
}

type approvalSignal struct {
	approve bool
	comment string
}

// EngineApp 全局单例(沿用项目 global 单例风格, 参考 TimedTaskServiceApp)
var EngineApp = NewEngine()

func NewEngine() *Engine {
	return &Engine{
		executor: newDefaultExecutor(),
		approval: make(map[uint]chan approvalSignal),
	}
}

// SetExecutor 替换执行器(主要供测试注入 fake executor; 生产环境用默认 http/shell 分派)
func (e *Engine) SetExecutor(ex StepExecutor) {
	if ex != nil {
		e.executor = ex
	}
}

// Run 执行一次构建(在独立 goroutine 中调用)
// triggerBy 用于 SSE 投递目标(投递给触发者)。构建状态由 DB 驱动, 取消靠轮询 status。
func (e *Engine) Run(ctx context.Context, buildID uint, triggerBy uint) {
	defer func() {
		if r := recover(); r != nil {
			logger.Bg().Mod("workflow").Error(fmt.Sprintf("build %d panic: %v", buildID, r))
			e.failBuild(buildID, fmt.Sprintf("引擎异常: %v", r), triggerBy)
		}
	}()

	// 加载 build + stage + step 运行实例(按快照顺序)
	detail, err := e.loadBuild(ctx, buildID)
	if err != nil {
		logErr(buildID, "加载构建实例失败", err)
		return
	}

	// SSE 投递目标: 触发者在线则推(离线静默, 状态以 DB 为准, 前端重连后拉取)
	publish := func(name string, payload interface{}) {
		e.publishEvent(triggerBy, name, payload)
	}

	// 解析 build.Params 为 map, 透传给执行器做 ${param.xxx} 变量替换
	params := decodeBuildParams(detail.Build.Params)

	publish("build:status", map[string]any{"buildId": buildID, "status": detail.Build.Status})

	stageFailed := false
	for i := range detail.Stages {
		st := &detail.Stages[i]

		// 取消检测
		if e.isCanceled(buildID) {
			e.finishBuild(buildID, modelWorkflow.BuildStatusCanceled, "已取消", triggerBy)
			return
		}

		// 标记 stage 开始
		now := time.Now()
		st.Status = modelWorkflow.BuildStatusRunning
		st.StartedAt = &now
		e.updateStage(st)
		publish("stage:status", st)

		// 执行该 stage 下所有 step
		// 串行(默认): 顺序跑, 失败即终止当前 stage 后续 step
		// 并行(Stage.Parallel=true): 并发跑所有 step, 收集失败数
		stageParallel := e.stageParallelMode(st.StageID)
		stepFailed := false
		if stageParallel {
			stepFailed = e.runStepsParallel(ctx, buildID, st.ID, detail.Steps, params, publish)
		} else {
			for j := range detail.Steps {
				sp := &detail.Steps[j]
				if sp.StageID != st.ID {
					continue
				}
				if e.isCanceled(buildID) {
					e.finishBuild(buildID, modelWorkflow.BuildStatusCanceled, "已取消", triggerBy)
					return
				}
				res := e.runStep(ctx, buildID, sp, params, publish)
				if res.Err != nil {
					stepFailed = true
					break // step 失败即终止当前 stage 后续 step
				}
			}
		}

		// 标记 stage 结束
		end := time.Now()
		st.FinishedAt = &end
		if stepFailed {
			st.Status = modelWorkflow.BuildStatusFailed
			e.updateStage(st)
			publish("stage:status", st)

			// 查定义: 该 stage 是否 ContinueOnError
			if !e.stageContinueOnError(st.StageID) {
				stageFailed = true
				break
			}
			// continue: 继续下一个 stage
			continue
		}
		st.Status = modelWorkflow.BuildStatusSuccess
		e.updateStage(st)
		publish("stage:status", st)

		// 审批 gate: Approval=true 的 stage 跑完后, 阻塞等待人工唤醒
		if e.stageNeedsApproval(st.StageID) {
			// 还有后续 stage 才需要等审批; 最后一个 stage 无需等
			if i < len(detail.Stages)-1 {
				// 先建审批通道, 再置状态, 避免 ApproveStage 抢先时丢信号
				_ = e.getOrCreateApprovalCh(buildID, true)
				e.setBuildStatus(buildID, modelWorkflow.BuildStatusApproval)
				publish("build:status", map[string]any{"buildId": buildID, "status": modelWorkflow.BuildStatusApproval})
				sig, ok := e.waitApproval(buildID)
				if !ok {
					// 引擎被清理/异常
					e.finishBuild(buildID, modelWorkflow.BuildStatusFailed, "审批通道异常", triggerBy)
					return
				}
				if !sig.approve {
					msg := "审批被拒绝"
					if sig.comment != "" {
						msg += ": " + sig.comment
					}
					e.finishBuild(buildID, modelWorkflow.BuildStatusFailed, msg, triggerBy)
					return
				}
				// 批准: 继续, 恢复 running; 清掉旧通道以便下次审批复用
				e.cleanupApproval(buildID)
				e.setBuildStatus(buildID, modelWorkflow.BuildStatusRunning)
				publish("build:status", map[string]any{"buildId": buildID, "status": modelWorkflow.BuildStatusRunning})
			}
		}
	}

	// 收尾
	if stageFailed {
		e.finishBuild(buildID, modelWorkflow.BuildStatusFailed, "存在阶段失败", triggerBy)
		return
	}
	e.finishBuild(buildID, modelWorkflow.BuildStatusSuccess, "成功", triggerBy)
}

// runStep 执行单个 step: 状态机 pending→running→success/failed, 日志落库+SSE
// params 透传给执行器做 ${param.xxx} 变量替换
func (e *Engine) runStep(ctx context.Context, buildID uint, sp *modelWorkflow.PipelineBuildStep, params map[string]string, publish func(string, any)) StepResult {
	now := time.Now()
	sp.Status = modelWorkflow.BuildStatusRunning
	sp.StartedAt = &now
	e.updateStep(sp)
	publish("step:status", sp)

	// 取该 step 的定义配置
	config := e.loadStepConfig(sp.StepID)

	// 日志 seq 自增 + 落库 + SSE
	seq := 0
	logFn := func(stream string, text string) {
		seq++
		e.appendLog(buildID, sp.ID, seq, stream, text)
		publish("step:log", map[string]any{
			"buildId": buildID,
			"stepId":  sp.ID,
			"seq":     seq,
			"stream":  stream,
			"text":    text,
			"ts":      time.Now().Format(time.RFC3339),
		})
	}

	res := e.executor.Execute(ctx, sp.SnapshotType, config, params, logFn)
	end := time.Now()
	sp.FinishedAt = &end
	sp.ExitCode = &res.ExitCode
	if res.Err != nil {
		sp.Status = modelWorkflow.BuildStatusFailed
		logFn(modelWorkflow.LogStreamSystem, "步骤失败: "+res.Err.Error())
	} else {
		sp.Status = modelWorkflow.BuildStatusSuccess
	}
	e.updateStep(sp)
	publish("step:status", sp)
	return res
}

// ============================== 数据访问(引擎内部, 简单封装) ==============================

func (e *Engine) loadBuild(ctx context.Context, buildID uint) (detail struct {
	Build  modelWorkflow.PipelineBuild
	Stages []modelWorkflow.PipelineBuildStage
	Steps  []modelWorkflow.PipelineBuildStep
}, err error) {
	err = global.GVA_DB.WithContext(ctx).First(&detail.Build, buildID).Error
	if err != nil {
		return
	}
	err = global.GVA_DB.WithContext(ctx).Where("build_id = ?", buildID).
		Order("snapshot_order ASC, id ASC").Find(&detail.Stages).Error
	if err != nil {
		return
	}
	err = global.GVA_DB.WithContext(ctx).Where("build_id = ?", buildID).
		Order("snapshot_order ASC, id ASC").Find(&detail.Steps).Error
	return
}

func (e *Engine) updateStage(st *modelWorkflow.PipelineBuildStage) {
	global.GVA_DB.Model(&modelWorkflow.PipelineBuildStage{}).Where("id = ?", st.ID).
		Updates(map[string]any{"status": st.Status, "started_at": st.StartedAt, "finished_at": st.FinishedAt})
}

func (e *Engine) updateStep(sp *modelWorkflow.PipelineBuildStep) {
	global.GVA_DB.Model(&modelWorkflow.PipelineBuildStep{}).Where("id = ?", sp.ID).
		Updates(map[string]any{
			"status":      sp.Status,
			"exit_code":   sp.ExitCode,
			"started_at":  sp.StartedAt,
			"finished_at": sp.FinishedAt,
		})
}

func (e *Engine) appendLog(buildID, stepID uint, seq int, stream, text string) {
	log := modelWorkflow.PipelineBuildLog{
		BuildID: buildID,
		StepID:  stepID,
		Seq:     seq,
		Stream:  stream,
		Text:    text,
		Ts:      time.Now(),
	}
	global.GVA_DB.Create(&log)
}

func (e *Engine) loadStepConfig(stepDefID uint) []byte {
	var step modelWorkflow.PipelineStep
	if err := global.GVA_DB.First(&step, stepDefID).Error; err != nil {
		return nil
	}
	return step.Config
}

func (e *Engine) stageNeedsApproval(stageDefID uint) bool {
	var st modelWorkflow.PipelineStage
	if err := global.GVA_DB.Select("approval").First(&st, stageDefID).Error; err != nil {
		return false
	}
	return st.Approval
}

func (e *Engine) stageContinueOnError(stageDefID uint) bool {
	var st modelWorkflow.PipelineStage
	if err := global.GVA_DB.Select("continue_on_error").First(&st, stageDefID).Error; err != nil {
		return false
	}
	return st.ContinueOnError
}

// stageParallelMode 查定义: 该 stage 内 step 是否并行执行
func (e *Engine) stageParallelMode(stageDefID uint) bool {
	var st modelWorkflow.PipelineStage
	if err := global.GVA_DB.Select("parallel").First(&st, stageDefID).Error; err != nil {
		return false
	}
	return st.Parallel
}

// runStepsParallel 并发执行某 stage 下所有 step, 返回是否有失败。
// 并行模式下所有 step 都会跑完(无法像串行那样中途 break); 失败与否由调用方按
// ContinueOnError 决定。日志 seq 由 runStep 内局部变量保证单 step 内有序,
// 跨 step 的日志交错是并行模式的固有行为(每条日志带 stepId 可区分)。
func (e *Engine) runStepsParallel(ctx context.Context, buildID, stageID uint, steps []modelWorkflow.PipelineBuildStep, params map[string]string, publish func(string, any)) bool {
	var wg sync.WaitGroup
	var failed int32
	for j := range steps {
		sp := &steps[j]
		if sp.StageID != stageID {
			continue
		}
		wg.Add(1)
		go func(s *modelWorkflow.PipelineBuildStep) {
			defer wg.Done()
			res := e.runStep(ctx, buildID, s, params, publish)
			if res.Err != nil {
				atomic.AddInt32(&failed, 1)
			}
		}(sp)
	}
	wg.Wait()
	return failed > 0
}

func (e *Engine) isCanceled(buildID uint) bool {
	var b modelWorkflow.PipelineBuild
	if err := global.GVA_DB.Select("status").First(&b, buildID).Error; err != nil {
		return false
	}
	return b.Status == modelWorkflow.BuildStatusCanceled
}

func (e *Engine) setBuildStatus(buildID uint, status string) {
	global.GVA_DB.Model(&modelWorkflow.PipelineBuild{}).Where("id = ?", buildID).
		Update("status", status)
}

func (e *Engine) finishBuild(buildID uint, status, msg string, triggerBy uint) {
	now := time.Now()
	global.GVA_DB.Model(&modelWorkflow.PipelineBuild{}).Where("id = ?", buildID).
		Updates(map[string]any{
			"status":      status,
			"message":     msg,
			"finished_at": now,
		})
	e.publishEvent(triggerBy, "build:status", map[string]any{"buildId": buildID, "status": status, "message": msg})
	// 失败/取消时, 向管理员角色(888) SSE 定向告警(离线静默丢弃, 不阻塞)
	if status == modelWorkflow.BuildStatusFailed || status == modelWorkflow.BuildStatusCanceled {
		e.alertFailure(buildID, status, msg)
	}
	e.cleanupApproval(buildID)
}

// alertFailure 失败告警: 查 888 角色用户, 经本体 SSE Hub 定向推送(仿 timedTask alertFailure)
func (e *Engine) alertFailure(buildID uint, status, msg string) {
	const alertAuthorityID = 888
	var ids []uint
	if err := global.GVA_DB.Table("sys_user_authority").
		Where("sys_authority_authority_id = ?", alertAuthorityID).
		Pluck("sys_user_id", &ids).Error; err != nil {
		return
	}
	if len(ids) == 0 {
		return
	}
	payload, _ := json.Marshal(map[string]any{
		"buildId": buildID,
		"status":  status,
		"error":   msg,
		"time":    time.Now().Format(time.RFC3339),
	})
	sse.Default().PublishToUsers(ids, sse.Event{Name: "workflow:alert", Data: string(payload)})
}

func (e *Engine) failBuild(buildID uint, msg string, triggerBy uint) {
	e.finishBuild(buildID, modelWorkflow.BuildStatusFailed, msg, triggerBy)
}

// ============================== SSE 推送 ==============================

func (e *Engine) publishEvent(userID uint, name string, payload interface{}) {
	if userID == 0 {
		return
	}
	data, err := json.Marshal(payload)
	if err != nil {
		return
	}
	sse.Default().Publish(userID, sse.Event{Name: name, Data: string(data)})
}

// ============================== 审批 gate ==============================

func (e *Engine) waitApproval(buildID uint) (approvalSignal, bool) {
	ch := e.getOrCreateApprovalCh(buildID, false)
	if ch == nil {
		return approvalSignal{}, false
	}
	sig, ok := <-ch
	return sig, ok
}

// NotifyApproval 由 BuildService.ApproveStage 调用: 唤醒等待中的引擎
func (e *Engine) NotifyApproval(buildID uint, approve bool, comment string) {
	ch := e.getOrCreateApprovalCh(buildID, true)
	if ch != nil {
		select {
		case ch <- approvalSignal{approve: approve, comment: comment}:
		default:
		}
	}
}

// getOrCreateApprovalCh 取或建审批通道; create=true 时若不存在则创建
func (e *Engine) getOrCreateApprovalCh(buildID uint, create bool) chan approvalSignal {
	e.amu.Lock()
	defer e.amu.Unlock()
	ch, ok := e.approval[buildID]
	if !ok && create {
		ch = make(chan approvalSignal, 1)
		e.approval[buildID] = ch
	}
	return ch
}

func (e *Engine) cleanupApproval(buildID uint) {
	e.amu.Lock()
	defer e.amu.Unlock()
	delete(e.approval, buildID)
}
