package service

import (
	"github.com/flipped-aurora/gin-vue-admin/server/service/example"
	"github.com/flipped-aurora/gin-vue-admin/server/service/media"
	"github.com/flipped-aurora/gin-vue-admin/server/service/ops"
	"github.com/flipped-aurora/gin-vue-admin/server/service/system"
	"github.com/flipped-aurora/gin-vue-admin/server/service/workflow"
)

var ServiceGroupApp = new(ServiceGroup)

type ServiceGroup struct {
	SystemServiceGroup   system.ServiceGroup
	ExampleServiceGroup  example.ServiceGroup
	MediaServiceGroup    media.ServiceGroup
	WorkflowServiceGroup workflow.ServiceGroup
	OpsServiceGroup      ops.ServiceGroup
}

func init() {
	// 注入流水线触发能力给运维工单审批: BuildService 实现了 ops.WorkflowTrigger 接口。
	// 放在 init 以保证包加载顺序无关(两边均完成 new 后赋值)。
	ops.WorkflowTriggerProvider = &ServiceGroupApp.WorkflowServiceGroup.BuildService
}
