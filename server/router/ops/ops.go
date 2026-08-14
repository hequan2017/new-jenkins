package ops

import (
	"github.com/flipped-aurora/gin-vue-admin/server/middleware"
	"github.com/gin-gonic/gin"
)

type OpsRouter struct{}

// InitOpsRouter 注册运维模块路由。
func (r *OpsRouter) InitOpsRouter(Router *gin.RouterGroup) {
	opsRouter := Router.Group("ops").Use(middleware.OperationRecord())
	opsRouterWithoutRecord := Router.Group("ops")
	{
		// 资产管理
		opsRouter.POST("createAsset", opsApi.CreateAsset)
		opsRouter.PUT("updateAsset", opsApi.UpdateAsset)
		opsRouter.DELETE("deleteAsset", opsApi.DeleteAsset)
		// 资产分组
		opsRouter.POST("createAssetGroup", opsApi.CreateAssetGroup)
		opsRouter.PUT("updateAssetGroup", opsApi.UpdateAssetGroup)
		opsRouter.DELETE("deleteAssetGroup", opsApi.DeleteAssetGroup)
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
		// 巡检任务
		opsRouter.POST("createInspectTask", opsApi.CreateInspectTask)
		opsRouter.PUT("updateInspectTask", opsApi.UpdateInspectTask)
		opsRouter.POST("toggleInspectTask", opsApi.ToggleInspectTask)
		opsRouter.POST("runInspectTask", opsApi.RunInspectTask)
		opsRouter.DELETE("deleteInspectTask", opsApi.DeleteInspectTask)
		// 备份任务
		opsRouter.POST("createBackupTask", opsApi.CreateBackupTask)
		opsRouter.PUT("updateBackupTask", opsApi.UpdateBackupTask)
		opsRouter.POST("toggleBackupTask", opsApi.ToggleBackupTask)
		opsRouter.POST("runBackupTask", opsApi.RunBackupTask)
		opsRouter.DELETE("deleteBackupTask", opsApi.DeleteBackupTask)
		// 告警处理
		opsRouter.POST("handleAlert", opsApi.HandleAlert)
		// 文件管理
		opsRouter.POST("writeFile", opsApi.WriteFile)
		opsRouter.POST("removeFile", opsApi.RemoveFile)
		opsRouter.POST("renameFile", opsApi.RenameFile)
		opsRouter.POST("mkdir", opsApi.Mkdir)
	}
	{
		opsRouterWithoutRecord.POST("getAssetList", opsApi.GetAssetList)
		opsRouterWithoutRecord.POST("getAssetGroupList", opsApi.GetAssetGroupList)
		opsRouterWithoutRecord.GET("getAllAssetGroups", opsApi.GetAllAssetGroups)
		opsRouterWithoutRecord.POST("getCredentialList", opsApi.GetCredentialList)
		opsRouterWithoutRecord.POST("getTicketList", opsApi.GetTicketList)
		opsRouterWithoutRecord.GET("findTicket", opsApi.FindTicket)
		opsRouterWithoutRecord.POST("getAuditList", opsApi.GetAuditList)
		opsRouterWithoutRecord.POST("getInspectTaskList", opsApi.GetInspectTaskList)
		opsRouterWithoutRecord.POST("getInspectResultList", opsApi.GetInspectResultList)
		opsRouterWithoutRecord.POST("getBackupTaskList", opsApi.GetBackupTaskList)
		opsRouterWithoutRecord.POST("getBackupRecordList", opsApi.GetBackupRecordList)
		opsRouterWithoutRecord.POST("getAlertList", opsApi.GetAlertList)
		opsRouterWithoutRecord.GET("getScheduleList", opsApi.GetScheduleList)
		opsRouterWithoutRecord.GET("getDashboard", opsApi.GetDashboard)
		// 列目录/读文件为查询性质, 不记操作记录
		opsRouterWithoutRecord.POST("listDir", opsApi.ListDir)
		opsRouterWithoutRecord.POST("readFile", opsApi.ReadFile)
		// 下载备份
		opsRouterWithoutRecord.GET("downloadBackup", opsApi.DownloadBackup)
		// 终端 WS: 不走 OperationRecord, 也不可套 Timeout(长连接)
		opsRouterWithoutRecord.GET("terminal", opsApi.TerminalStream)
	}
}
