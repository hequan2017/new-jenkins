package ops

import (
	"context"

	"github.com/flipped-aurora/gin-vue-admin/server/model/workflow"
)

// WorkflowTrigger 抽象"触发流水线构建"能力, 避免运维包直接依赖顶层 service 包造成循环导入。
// 由顶层 service/enter.go 在初始化时注入 workflow.BuildService 的实现。
type WorkflowTrigger interface {
	TriggerBuild(ctx context.Context, pipelineID uint, params []workflow.ParamValue, trigger string, triggerBy uint) (buildID uint, err error)
}

// WorkflowTriggerProvider 包级注入点, 顶层 service 包初始化时赋值。
// 为 nil 时审批通过会返回错误, 提示未注入。
var WorkflowTriggerProvider WorkflowTrigger
