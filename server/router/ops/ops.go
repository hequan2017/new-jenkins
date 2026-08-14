package ops

import (
	"github.com/flipped-aurora/gin-vue-admin/server/middleware"
	"github.com/gin-gonic/gin"
)

type OpsRouter struct{}

// InitOpsRouter 注册运维模块路由(资产 / 凭据 / 跳板机 / 工单)。
// 全部挂在 PrivateGroup, 终端 WS 路由不挂 OperationRecord 与 Timeout 中间件。
func (r *OpsRouter) InitOpsRouter(Router *gin.RouterGroup) {
	opsRouter := Router.Group("ops").Use(middleware.OperationRecord())
	opsRouterWithoutRecord := Router.Group("ops")
	{
		// 资产管理
		opsRouter.POST("createAsset", opsApi.CreateAsset)
		opsRouter.PUT("updateAsset", opsApi.UpdateAsset)
		opsRouter.DELETE("deleteAsset", opsApi.DeleteAsset)
		// 凭据管理
		opsRouter.POST("createCredential", opsApi.CreateCredential)
		opsRouter.PUT("updateCredential", opsApi.UpdateCredential)
		opsRouter.DELETE("deleteCredential", opsApi.DeleteCredential)
		// 跳板机
		opsRouter.POST("testConnection", opsApi.TestConnection)
		opsRouter.POST("execCommand", opsApi.ExecCommand)
		// 工单发版
		opsRouter.POST("createTicket", opsApi.CreateTicket)
		opsRouter.DELETE("deleteTicket", opsApi.DeleteTicket)
		opsRouter.POST("approveTicket", opsApi.ApproveTicket)
		opsRouter.POST("cancelTicket", opsApi.CancelTicket)
	}
	{
		opsRouterWithoutRecord.POST("getAssetList", opsApi.GetAssetList)
		opsRouterWithoutRecord.POST("getCredentialList", opsApi.GetCredentialList)
		opsRouterWithoutRecord.POST("getTicketList", opsApi.GetTicketList)
		opsRouterWithoutRecord.GET("findTicket", opsApi.FindTicket)
		// 终端 WS: 不走 OperationRecord, 也不可套 Timeout(长连接)
		opsRouterWithoutRecord.GET("terminal", opsApi.TerminalStream)
	}
}
