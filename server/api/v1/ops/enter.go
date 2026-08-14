package ops

import "github.com/flipped-aurora/gin-vue-admin/server/service"

type ApiGroup struct {
	AssetApi
	AssetGroupApi
	CredentialApi
	BastionApi
	TicketApi
	AuditApi
	InspectApi
	BackupApi
	DashboardApi
	FileApi
	AlertApi
	ScheduleCenterApi
}

var (
	assetService          = service.ServiceGroupApp.OpsServiceGroup.AssetService
	assetGroupService     = service.ServiceGroupApp.OpsServiceGroup.AssetGroupService
	credentialService     = service.ServiceGroupApp.OpsServiceGroup.CredentialService
	sshService            = service.ServiceGroupApp.OpsServiceGroup.SSHService
	terminalService       = service.ServiceGroupApp.OpsServiceGroup.TerminalService
	ticketService         = service.ServiceGroupApp.OpsServiceGroup.TicketService
	auditService          = service.ServiceGroupApp.OpsServiceGroup.AuditService
	inspectService        = service.ServiceGroupApp.OpsServiceGroup.InspectService
	backupService         = service.ServiceGroupApp.OpsServiceGroup.BackupService
	dashboardService      = service.ServiceGroupApp.OpsServiceGroup.DashboardService
	fileService           = service.ServiceGroupApp.OpsServiceGroup.FileService
	alertService          = service.ServiceGroupApp.OpsServiceGroup.AlertService
	scheduleCenterService = service.ServiceGroupApp.OpsServiceGroup.ScheduleCenterService
)
