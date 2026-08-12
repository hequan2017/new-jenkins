package workflow

import (
	"time"

	"github.com/flipped-aurora/gin-vue-admin/server/global"
)

// PipelineBuildStage 一次构建中每个 Stage 的运行视图(对应 Jenkins Stage 视图的一个格子)
// 与 PipelineStage(定义)分离:定义可被修改, 历史构建仍保留当时的执行记录。
// SnapshotName/SnapshotOrder 在构建启动时从定义拷贝, 之后不受定义变更影响。
type PipelineBuildStage struct {
	global.GVA_MODEL
	BuildID     uint       `json:"buildId" form:"buildId" gorm:"index;column:build_id;comment:所属构建ID"`
	StageID     uint       `json:"stageId" form:"stageId" gorm:"column:stage_id;comment:对应阶段定义ID"`
	SnapshotName  string   `json:"snapshotName" form:"snapshotName" gorm:"column:snapshot_name;comment:构建时的阶段名快照"`
	SnapshotOrder int      `json:"snapshotOrder" form:"snapshotOrder" gorm:"column:snapshot_order;comment:构建时的阶段顺序快照"`
	Status      string     `json:"status" form:"status" gorm:"column:status;comment:阶段状态"`
	StartedAt   *time.Time `json:"startedAt" form:"startedAt" gorm:"column:started_at;comment:开始时间"`
	FinishedAt  *time.Time `json:"finishedAt" form:"finishedAt" gorm:"column:finished_at;comment:结束时间"`
}

func (PipelineBuildStage) TableName() string {
	return "wf_pipeline_build_stages"
}
