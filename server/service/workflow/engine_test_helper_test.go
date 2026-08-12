package workflow

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/flipped-aurora/gin-vue-admin/server/global"
	"github.com/flipped-aurora/gin-vue-admin/server/internal/testutil"
	modelWorkflow "github.com/flipped-aurora/gin-vue-admin/server/model/workflow"
)

var memDBCounter uint32

// initMemoryDB 建共享内存库 + AutoMigrate workflow 全部表 + 赋值 global.GVA_DB(并 cleanup 还原)
// 注意: Engine 测试跨 goroutine 读写同一内存库。glebarez/sqlite 底层为单连接池, 多 goroutine
// 共享同一份 :memory: 数据; 但单连接会在审批等待 goroutine 占用时阻塞查询, 因此审批测试
// 用 NotifyApproval 唤醒后 Run 很快释放连接, 测试侧轮询容忍短时阻塞。
func initMemoryDB(t *testing.T) error {
	_ = testutil.NewMemoryDB(t,
		&modelWorkflow.Pipeline{}, &modelWorkflow.PipelineStage{}, &modelWorkflow.PipelineStep{},
		&modelWorkflow.PipelineBuild{}, &modelWorkflow.PipelineBuildStage{},
		&modelWorkflow.PipelineBuildStep{}, &modelWorkflow.PipelineBuildLog{},
	)
	_ = testutil.InitNopLogger()
	return nil
}

// gvaCreate / gvaFirst: 测试内直接走 global.GVA_DB, 不走 service 层(避免循环)
func gvaCreate(v interface{}) error { return global.GVA_DB.Create(v).Error }

func gvaFirst(out interface{}, id uint) error {
	return global.GVA_DB.First(out, id).Error
}

func ptrTime(t time.Time) *time.Time { return &t }

func jsonBytes(s string) []byte { return []byte(s) }

// createPipelineForTest 直接落库 pipeline + stages + steps, 并把生成的定义 ID 回填到入参切片
// (build step seed 需要正确的 step 定义 ID)。用指针/索引操作, 避免 Create 回填落到拷贝上。
func createPipelineForTest(p *modelWorkflow.Pipeline, stages []modelWorkflow.PipelineStage) error {
	p.Stages = nil
	if err := global.GVA_DB.Create(p).Error; err != nil {
		return err
	}
	for i := range stages {
		stg := &stages[i]
		stg.PipelineID = p.ID
		steps := stg.Steps
		stg.Steps = nil
		if err := global.GVA_DB.Create(stg).Error; err != nil {
			return err
		} // stg.ID 已回填
		// 重建 steps 到入参切片并落库, 用索引操作让回填生效
		stg.Steps = make([]modelWorkflow.PipelineStep, len(steps))
		for j := range steps {
			stg.Steps[j] = steps[j]
			stg.Steps[j].StageID = stg.ID
			if !json.Valid(stg.Steps[j].Config) {
				stg.Steps[j].Config = nil
			}
			if err := global.GVA_DB.Create(&stg.Steps[j]).Error; err != nil {
				return err
			} // Steps[j].ID 已回填
		}
	}
	return nil
}
