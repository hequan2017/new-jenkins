package workflow

import (
	"time"

	"github.com/flipped-aurora/gin-vue-admin/server/global"
	"gorm.io/datatypes"
)

// PipelineBuildStep 一次构建中每个 Step 的运行视图(对应 Jenkins 一个 Step 的执行记录)
// 日志正文不入本表(避免行膨胀), 单独存 PipelineBuildLog。
type PipelineBuildStep struct {
	global.GVA_MODEL
	BuildID        uint           `json:"buildId" form:"buildId" gorm:"index;column:build_id;comment:所属构建ID"`
	StageID        uint           `json:"stageId" form:"stageId" gorm:"index;column:stage_id;comment:所属构建阶段ID"`
	StepID         uint           `json:"stepId" form:"stepId" gorm:"column:step_id;comment:对应步骤定义ID"`
	SnapshotName   string         `json:"snapshotName" form:"snapshotName" gorm:"column:snapshot_name;comment:构建时的步骤名快照"`
	SnapshotType   string         `json:"snapshotType" form:"snapshotType" gorm:"column:snapshot_type;comment:构建时的步骤类型快照"`
	SnapshotConfig datatypes.JSON `json:"snapshotConfig" form:"snapshotConfig" gorm:"column:snapshot_config;comment:构建时的步骤配置快照" swaggertype:"object"`
	SnapshotOrder  int            `json:"snapshotOrder" form:"snapshotOrder" gorm:"column:snapshot_order;comment:构建时的步骤顺序快照"`
	Status         string         `json:"status" form:"status" gorm:"column:status;comment:步骤状态"`
	ExitCode       *int           `json:"exitCode" form:"exitCode" gorm:"column:exit_code;comment:退出码(shell 步骤)"`
	StartedAt      *time.Time     `json:"startedAt" form:"startedAt" gorm:"column:started_at;comment:开始时间"`
	FinishedAt     *time.Time     `json:"finishedAt" form:"finishedAt" gorm:"column:finished_at;comment:结束时间"`
}

func (PipelineBuildStep) TableName() string {
	return "wf_pipeline_build_steps"
}
