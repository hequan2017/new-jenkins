package initialize

import (
	"github.com/flipped-aurora/gin-vue-admin/server/global"
	"github.com/flipped-aurora/gin-vue-admin/server/model/ops"
	"github.com/flipped-aurora/gin-vue-admin/server/model/workflow"
)

// bizModel 注册业务模块的表结构(与 system 核心表分离, 挂业务模块用)
func bizModel() error {
	db := global.GVA_DB
	err := db.AutoMigrate(
		// workflow 流水线引擎
		workflow.Pipeline{},
		workflow.PipelineStage{},
		workflow.PipelineStep{},
		workflow.PipelineBuild{},
		workflow.PipelineBuildStage{},
		workflow.PipelineBuildStep{},
		workflow.PipelineBuildLog{},
		// ops 运维模块(资产管理 / 跳板机 / 工单发版 / 审计 / 巡检)
		ops.Asset{},
		ops.Credential{},
		ops.Ticket{},
		ops.AuditRecord{},
		ops.InspectTask{},
		ops.InspectResult{},
	)
	if err != nil {
		return err
	}
	return nil
}
