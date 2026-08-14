package ops

import "github.com/flipped-aurora/gin-vue-admin/server/service"

type ApiGroup struct {
	AssetApi
	CredentialApi
	BastionApi
	TicketApi
}

var (
	assetService      = service.ServiceGroupApp.OpsServiceGroup.AssetService
	credentialService = service.ServiceGroupApp.OpsServiceGroup.CredentialService
	sshService        = service.ServiceGroupApp.OpsServiceGroup.SSHService
	terminalService   = service.ServiceGroupApp.OpsServiceGroup.TerminalService
	ticketService     = service.ServiceGroupApp.OpsServiceGroup.TicketService
)
