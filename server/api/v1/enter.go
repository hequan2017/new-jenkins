package v1

import (
	"github.com/flipped-aurora/gin-vue-admin/server/api/v1/example"
	"github.com/flipped-aurora/gin-vue-admin/server/api/v1/media"
	"github.com/flipped-aurora/gin-vue-admin/server/api/v1/ops"
	"github.com/flipped-aurora/gin-vue-admin/server/api/v1/system"
	"github.com/flipped-aurora/gin-vue-admin/server/api/v1/workflow"
)

var ApiGroupApp = new(ApiGroup)

type ApiGroup struct {
	SystemApiGroup   system.ApiGroup
	ExampleApiGroup  example.ApiGroup
	MediaApiGroup    media.ApiGroup
	WorkflowApiGroup workflow.ApiGroup
	OpsApiGroup      ops.ApiGroup
}
