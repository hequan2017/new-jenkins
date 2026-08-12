package workflow

import "github.com/flipped-aurora/gin-vue-admin/server/service"

type ApiGroup struct {
	WorkflowApi
}

var (
	pipelineService = service.ServiceGroupApp.WorkflowServiceGroup.PipelineService
	buildService    = service.ServiceGroupApp.WorkflowServiceGroup.BuildService
)
