// server/initialize/workflow_schedule.go
package initialize

import (
	"context"

	"github.com/flipped-aurora/gin-vue-admin/server/global"
	"github.com/flipped-aurora/gin-vue-admin/server/service"
	"github.com/flipped-aurora/gin-vue-admin/server/utils/datascope"
	"github.com/flipped-aurora/gin-vue-admin/server/utils/logger"
)

// LoadWorkflowSchedules 从 wf_pipelines 恢复 schedule 类型流水线的定时调度(幂等)。
// 必须在 RegisterTables(建表)之后调用。
func LoadWorkflowSchedules() {
	if global.GVA_DB == nil {
		return
	}
	ctx := datascope.WithSystem(context.Background())
	if err := service.ServiceGroupApp.WorkflowServiceGroup.LoadAll(ctx); err != nil {
		logger.Bg().Mod("workflow").Err(err).Error("流水线定时调度启动加载失败")
	}
}
