package workflow

import api "github.com/flipped-aurora/gin-vue-admin/server/api/v1"

type RouterGroup struct {
	WorkflowRouter
}

var (
	workflowApi = api.ApiGroupApp.WorkflowApiGroup.WorkflowApi
)
