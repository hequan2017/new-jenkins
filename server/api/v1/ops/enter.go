package ops

import "github.com/flipped-aurora/gin-vue-admin/server/service"

type ApiGroup struct {
	AssetApi
	CredentialApi
	BastionApi
	TicketApi
	AuditApi
	InspectApi
	DashboardApi
	FileApi
}

var (
	assetService      = service.ServiceGroupApp.OpsServiceGroup.AssetService
	credentialService = service.ServiceGroupApp.OpsServiceGroup.CredentialService
	sshService        = service.ServiceGroupApp.OpsServiceGroup.SSHService
	terminalService   = service.ServiceGroupApp.OpsServiceGroup.TerminalService
	ticketService     = service.ServiceGroupApp.OpsServiceGroup.TicketService
	auditService      = service.ServiceGroupApp.OpsServiceGroup.AuditService
	inspectService    = service.ServiceGroupApp.OpsServiceGroup.InspectService
	dashboardService  = service.ServiceGroupApp.OpsServiceGroup.DashboardService
	fileService       = service.ServiceGroupApp.OpsServiceGroup.FileService
)
