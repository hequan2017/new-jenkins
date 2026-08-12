package workflow

import (
	"time"

	"github.com/flipped-aurora/gin-vue-admin/server/global"
	"gorm.io/datatypes"
)

// 构建状态机
const (
	BuildStatusPending  = "pending"          // 已创建未开始(理论中间态, 通常直接进 running)
	BuildStatusRunning  = "running"          // 执行中
	BuildStatusApproval = "running-approval" // 等待人工审批 gate
	BuildStatusSuccess  = "success"
	BuildStatusFailed   = "failed"
	BuildStatusCanceled = "canceled"
)

// PipelineBuild 一次流水线执行实例(对应 Jenkins 的 Build / Run)
// Status 取 BuildStatus* 常量。TriggerBy 为触发人用户ID(普通字段, 非数据权限归属列)。
type PipelineBuild struct {
	global.GVA_MODEL
	PipelineID uint           `json:"pipelineId" form:"pipelineId" gorm:"index;uniqueIndex:idx_wf_pipeline_build_no;column:pipeline_id;comment:所属流水线ID"`
	BuildNo    int            `json:"buildNo" form:"buildNo" gorm:"uniqueIndex:idx_wf_pipeline_build_no;column:build_no;comment:构建序号(同流水线下自增)"`
	Status     string         `json:"status" form:"status" gorm:"index;column:status;comment:构建状态"`
	Params     datatypes.JSON `json:"params" form:"params" gorm:"column:params;comment:本次构建入参(JSON)" swaggertype:"object"`
	Trigger    string         `json:"trigger" form:"trigger" gorm:"column:trigger;comment:触发方式 manual|schedule|webhook"`
	TriggerBy  uint           `json:"triggerBy" form:"triggerBy" gorm:"column:trigger_by;comment:触发人用户ID"`
	Message    string         `json:"message" form:"message" gorm:"column:message;comment:构建结果或失败原因摘要"`
	StartedAt  *time.Time     `json:"startedAt" form:"startedAt" gorm:"column:started_at;comment:开始时间"`
	FinishedAt *time.Time     `json:"finishedAt" form:"finishedAt" gorm:"column:finished_at;comment:结束时间"`
}

func (PipelineBuild) TableName() string {
	return "wf_pipeline_builds"
}
