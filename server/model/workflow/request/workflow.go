package request

import (
	common "github.com/flipped-aurora/gin-vue-admin/server/model/common/request"

	"github.com/flipped-aurora/gin-vue-admin/server/model/workflow"
)

// CreatePipelineReq 创建流水线请求(含 Stage / Step 树, 一次性级联创建)
// 更新走同一棵树: 全量覆盖式保存。
type CreatePipelineReq struct {
	Name        string                  `json:"name" form:"name"`
	Description string                  `json:"description" form:"description"`
	TriggerType string                  `json:"triggerType" form:"triggerType"`
	Enabled     bool                    `json:"enabled" form:"enabled"`
	ParamSchema []workflow.ParamField   `json:"paramSchema" form:"paramSchema"`
	Stages      []workflow.PipelineStage `json:"stages" form:"stages"`
}

// PipelineSearch 流水线分页查询
type PipelineSearch struct {
	common.PageInfo
	Name        string `json:"name" form:"name"`
	TriggerType string `json:"triggerType" form:"triggerType"`
	Enabled     *bool  `json:"enabled" form:"enabled"`
}

// PipelineBuildSearch 构建分页查询
type PipelineBuildSearch struct {
	common.PageInfo
	PipelineID uint   `json:"pipelineId" form:"pipelineId"`
	Status     string `json:"status" form:"status"`
	Trigger    string `json:"trigger" form:"trigger"`
}

// TriggerBuildReq 触发构建
type TriggerBuildReq struct {
	PipelineID uint                    `json:"pipelineId" form:"pipelineId"`
	Params     []workflow.ParamValue   `json:"params" form:"params"`
}

// ApproveStageReq 审批 gate:批准 / 拒绝
type ApproveStageReq struct {
	BuildID uint   `json:"buildId" form:"buildId"`
	Approve bool   `json:"approve" form:"approve"` // true=通过继续下一阶段, false=拒绝标记失败
	Comment string `json:"comment" form:"comment"`
}

// BuildLogSearch 构建日志分页(按 Step 维度拉历史日志)
type BuildLogSearch struct {
	common.PageInfo
	BuildID uint `json:"buildId" form:"buildId"`
	StepID  uint `json:"stepId" form:"stepId"`
}
