package workflow

import (
	"github.com/flipped-aurora/gin-vue-admin/server/middleware"
	"github.com/gin-gonic/gin"
)

type WorkflowRouter struct{}

func (r *WorkflowRouter) InitWorkflowRouter(Router *gin.RouterGroup) {
	workflowRouter := Router.Group("workflow").Use(middleware.OperationRecord())
	workflowRouterWithoutRecord := Router.Group("workflow")
	{
		workflowRouter.POST("createPipeline", workflowApi.CreatePipeline)   // 创建流水线
		workflowRouter.PUT("updatePipeline", workflowApi.UpdatePipeline)    // 更新流水线
		workflowRouter.DELETE("deletePipeline", workflowApi.DeletePipeline) // 删除流水线
		workflowRouter.POST("triggerBuild", workflowApi.TriggerBuild)       // 触发构建
		workflowRouter.POST("cancelBuild", workflowApi.CancelBuild)         // 取消构建
		workflowRouter.POST("approveStage", workflowApi.ApproveStage)       // 审批 gate
	}
	{
		workflowRouterWithoutRecord.POST("getPipelineList", workflowApi.GetPipelineList)   // 流水线列表
		workflowRouterWithoutRecord.GET("findPipeline", workflowApi.FindPipeline)           // 流水线详情
		workflowRouterWithoutRecord.POST("getBuildList", workflowApi.GetBuildList)          // 构建历史
		workflowRouterWithoutRecord.GET("getBuildDetail", workflowApi.GetBuildDetail)       // 构建详情
		workflowRouterWithoutRecord.GET("getBuildLog", workflowApi.GetBuildLog)             // 构建日志
		// SSE 长连接: 绝不能套 TimeoutMiddleware(参考 utils/sse/hub.go:168 约束)
		workflowRouterWithoutRecord.GET("buildStream", workflowApi.BuildStream)             // 构建实时事件
	}
}
