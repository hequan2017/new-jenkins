package workflow

import (
	"encoding/json"
	"sync/atomic"
	"testing"
	"time"

	"github.com/flipped-aurora/gin-vue-admin/server/global"
	"github.com/flipped-aurora/gin-vue-admin/server/internal/testutil"
	modelWorkflow "github.com/flipped-aurora/gin-vue-admin/server/model/workflow"
	"github.com/glebarez/sqlite"
	"gorm.io/gorm"
)

var memDBCounter uint32

// initMemoryDB 建共享内存库 + AutoMigrate workflow 全部表 + 赋值 global.GVA_DB(并 cleanup 还原)
// 注意: Engine 测试跨 goroutine 读写同一内存库。glebarez 的 ":memory:" 默认每连接独立库,
// 跨 goroutine 会拿到空库。这里用 "file:<uniq>?mode=memory&cache=shared" 让多连接共享同一份
// 内存库, 避免单连接池在审批等待时被 Run goroutine 占用而阻塞主 goroutine 的轮询查询。
// 同时 migrate sys_user_authority 表(失败告警 alertFailure 会查它), 无该表会让 gorm 报错。
func initMemoryDB(t *testing.T) error {
	dsn := "file:wftest" + uitoa(atomic.AddUint32(&memDBCounter, 1)) + "?mode=memory&cache=shared"
	db, err := gorm.Open(sqlite.Open(dsn), &gorm.Config{})
	if err != nil {
		t.Fatalf("open shared memory db: %v", err)
	}
	if err := db.AutoMigrate(
		&modelWorkflow.Pipeline{}, &modelWorkflow.PipelineStage{}, &modelWorkflow.PipelineStep{},
		&modelWorkflow.PipelineBuild{}, &modelWorkflow.PipelineBuildStage{},
		&modelWorkflow.PipelineBuildStep{}, &modelWorkflow.PipelineBuildLog{},
	); err != nil {
		t.Fatalf("automigrate: %v", err)
	}
	// alertFailure 查询 sys_user_authority 表, 建空表避免 "no such table"
	if err := db.Exec(`CREATE TABLE IF NOT EXISTS sys_user_authority (
		sys_user_id bigint, sys_authority_authority_id bigint, PRIMARY KEY(sys_user_id, sys_authority_authority_id)
	)`).Error; err != nil {
		t.Fatalf("create sys_user_authority: %v", err)
	}
	old := global.GVA_DB
	global.GVA_DB = db
	t.Cleanup(func() { global.GVA_DB = old })
	_ = testutil.InitNopLogger()
	return nil
}

func uitoa(n uint32) string {
	if n == 0 {
		return "0"
	}
	buf := []byte{}
	for n > 0 {
		buf = append([]byte{byte('0' + n%10)}, buf...)
		n /= 10
	}
	return string(buf)
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
