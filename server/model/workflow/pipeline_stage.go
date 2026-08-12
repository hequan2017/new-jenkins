package workflow

import (
	"github.com/flipped-aurora/gin-vue-admin/server/global"
)

// PipelineStage 流水线阶段定义(对应 Jenkins 的 Stage)
// 同一 Pipeline 下按 Order 升序执行。Approval=true 的 Stage 跑完前置 Step 后,
// 会把 Build 置于 running-approval 状态, 等人工 approveStage 接口唤醒再继续下一 Stage。
type PipelineStage struct {
	global.GVA_MODEL
	PipelineID    uint   `json:"pipelineId" form:"pipelineId" gorm:"index;column:pipeline_id;comment:所属流水线ID"`
	Name          string `json:"name" form:"name" gorm:"column:name;comment:阶段名称"`
	Order         int    `json:"order" form:"order" gorm:"column:sort_order;comment:阶段顺序(升序)"`
	Approval      bool   `json:"approval" form:"approval" gorm:"column:approval;comment:该阶段是否需要人工审批 gate"`
	ContinueOnError bool `json:"continueOnError" form:"continueOnError" gorm:"column:continue_on_error;comment:阶段内 Step 失败后是否继续后续阶段"`
	// Steps 关联, 查询时 Preload; form:"-" 避免 query 绑定递归
	Steps []PipelineStep `json:"steps" form:"-" gorm:"foreignKey:StageID;references:ID;constraint:OnDelete:CASCADE"`
}

func (PipelineStage) TableName() string {
	return "wf_pipeline_stages"
}
