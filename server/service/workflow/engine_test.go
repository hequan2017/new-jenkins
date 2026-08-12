package workflow

import (
	"context"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	modelWorkflow "github.com/flipped-aurora/gin-vue-admin/server/model/workflow"
	"github.com/flipped-aurora/gin-vue-admin/server/utils/datascope"
)

// fakeExecutor 测试用可控执行器: 按 step 名决定成功/失败, 记录调用顺序
type fakeExecutor struct {
	mu          sync.Mutex
	calls       []string
	failOn      map[string]bool
	approvalHook func() // 执行到某 step 时阻塞, 用于测试审批 gate
}

func (f *fakeExecutor) Execute(ctx context.Context, stepType string, config []byte, log logFunc) StepResult {
	f.mu.Lock()
	f.calls = append(f.calls, string(config))
	name := string(config)
	f.mu.Unlock()
	log("stdout", "exec "+name)
	if f.failOn[name] {
		return StepResult{ExitCode: 1, Err: errFake(name)}
	}
	return StepResult{ExitCode: 0}
}

type errFake string

func (e errFake) Error() string { return "fake fail on " + string(e) }

// seedPipeline 构造一条流水线定义 + 运行实例, 返回 buildID
func seedPipeline(t *testing.T, stages []modelWorkflow.PipelineStage) uint {
	t.Helper()
	// 1. pipeline
	p := modelWorkflow.Pipeline{
		Name:        "test-pipeline",
		TriggerType: modelWorkflow.TriggerManual,
		Enabled:     true,
	}
	if err := initMemoryDB(t); err != nil {
		t.Fatalf("NewMemoryDB returned err: %v", err)
	}
	if err := createPipelineForTest(&p, stages); err != nil {
		t.Fatalf("seed pipeline: %v", err)
	}
	// 2. build + run 实例
	build := modelWorkflow.PipelineBuild{
		PipelineID: p.ID,
		BuildNo:    1,
		Status:     modelWorkflow.BuildStatusRunning,
		Trigger:    modelWorkflow.TriggerManual,
		StartedAt:  ptrTime(time.Now()),
	}
	if err := gvaCreate(&build); err != nil {
		t.Fatalf("seed build: %v", err)
	}
	for _, st := range stages {
		bs := modelWorkflow.PipelineBuildStage{
			BuildID: build.ID, StageID: st.ID,
			SnapshotName: st.Name, SnapshotOrder: st.Order,
			Status: modelWorkflow.BuildStatusPending,
		}
		if err := gvaCreate(&bs); err != nil {
			t.Fatalf("seed build stage: %v", err)
		}
		for _, sp := range st.Steps {
			if err := gvaCreate(&modelWorkflow.PipelineBuildStep{
				BuildID: build.ID, StageID: bs.ID, StepID: sp.ID,
				SnapshotName: sp.Name, SnapshotType: sp.Type, SnapshotOrder: sp.Order,
				Status: modelWorkflow.BuildStatusPending,
			}); err != nil {
				t.Fatalf("seed build step: %v", err)
			}
		}
	}
	return build.ID
}

// TestEngineSuccessPath 单 stage 两 step 全成功 -> build success
func TestEngineSuccessPath(t *testing.T) {
	stages := []modelWorkflow.PipelineStage{
		{
			Name: "build", Order: 1,
			Steps: []modelWorkflow.PipelineStep{
				{Name: "s1", Type: modelWorkflow.StepTypeShell, Config: jsonBytes(`"s1"`), Order: 1},
				{Name: "s2", Type: modelWorkflow.StepTypeShell, Config: jsonBytes(`"s2"`), Order: 2},
			},
		},
	}
	buildID := seedPipeline(t, stages)

	fake := &fakeExecutor{}
	EngineApp.SetExecutor(fake)
	defer EngineApp.SetExecutor(newDefaultExecutor())

	EngineApp.Run(datascope.WithSystem(context.Background()), buildID, 0)

	// 校验 build 状态
	var b modelWorkflow.PipelineBuild
	if err := gvaFirst(&b, buildID); err != nil {
		t.Fatal(err)
	}
	if b.Status != modelWorkflow.BuildStatusSuccess {
		t.Fatalf("期望 success, 实际 %s, msg=%s", b.Status, b.Message)
	}
	if len(fake.calls) != 2 {
		t.Fatalf("期望执行 2 个 step, 实际 %d", len(fake.calls))
	}
}

// TestEngineStepFailureBuildFailed step 失败且非 continueOnError -> build failed
func TestEngineStepFailureBuildFailed(t *testing.T) {
	stages := []modelWorkflow.PipelineStage{
		{
			Name: "stage-a", Order: 1,
			Steps: []modelWorkflow.PipelineStep{
				{Name: "ok", Type: modelWorkflow.StepTypeShell, Config: jsonBytes(`"ok"`), Order: 1},
				{Name: "boom", Type: modelWorkflow.StepTypeShell, Config: jsonBytes(`"boom"`), Order: 2},
			},
		},
		{
			Name:  "stage-b", Order: 2,
			Steps: []modelWorkflow.PipelineStep{
				{Name: "never", Type: modelWorkflow.StepTypeShell, Config: jsonBytes(`"never"`), Order: 1},
			},
		},
	}
	buildID := seedPipeline(t, stages)

	fake := &fakeExecutor{failOn: map[string]bool{`"boom"`: true}}
	EngineApp.SetExecutor(fake)
	defer EngineApp.SetExecutor(newDefaultExecutor())

	EngineApp.Run(datascope.WithSystem(context.Background()), buildID, 0)

	var b modelWorkflow.PipelineBuild
	if err := gvaFirst(&b, buildID); err != nil {
		t.Fatal(err)
	}
	if b.Status != modelWorkflow.BuildStatusFailed {
		t.Fatalf("期望 failed, 实际 %s", b.Status)
	}
	// stage-b 的 step 不应执行
	for _, c := range fake.calls {
		if c == `"never"` {
			t.Fatal("失败 stage 后续 step 不应执行")
		}
	}
}

// TestEngineApprovalGate 审批 gate: 第一 stage 跑完后等待审批, 批准后第二 stage 继续成功
func TestEngineApprovalGate(t *testing.T) {
	stages := []modelWorkflow.PipelineStage{
		{
			Name: "gate-stage", Order: 1, Approval: true,
			Steps: []modelWorkflow.PipelineStep{
				{Name: "pre", Type: modelWorkflow.StepTypeShell, Config: jsonBytes(`"pre"`), Order: 1},
			},
		},
		{
			Name: "post-stage", Order: 2,
			Steps: []modelWorkflow.PipelineStep{
				{Name: "post", Type: modelWorkflow.StepTypeShell, Config: jsonBytes(`"post"`), Order: 1},
			},
		},
	}
	buildID := seedPipeline(t, stages)

	var execCount int32
	fake := &fakeExecutor{}
	EngineApp.SetExecutor(fake)
	defer EngineApp.SetExecutor(newDefaultExecutor())

	done := make(chan struct{})
	go func() {
		EngineApp.Run(datascope.WithSystem(context.Background()), buildID, 0)
		atomic.AddInt32(&execCount, 1)
		close(done)
	}()

	// 轮询等 build 进入 approval
	deadline := time.Now().Add(3 * time.Second)
	for time.Now().Before(deadline) {
		var b modelWorkflow.PipelineBuild
		_ = gvaFirst(&b, buildID)
		if b.Status == modelWorkflow.BuildStatusApproval {
			break
		}
		time.Sleep(20 * time.Millisecond)
	}

	// 批准
	EngineApp.NotifyApproval(buildID, true, "")

	select {
	case <-done:
	case <-time.After(3 * time.Second):
		t.Fatal("引擎未在审批后完成")
	}

	var b modelWorkflow.PipelineBuild
	if err := gvaFirst(&b, buildID); err != nil {
		t.Fatal(err)
	}
	if b.Status != modelWorkflow.BuildStatusSuccess {
		t.Fatalf("期望审批通过后 success, 实际 %s msg=%s", b.Status, b.Message)
	}
	// post step 应被执行
	found := false
	for _, c := range fake.calls {
		if c == `"post"` {
			found = true
		}
	}
	if !found {
		t.Fatal("审批通过后 post step 应执行")
	}
}
