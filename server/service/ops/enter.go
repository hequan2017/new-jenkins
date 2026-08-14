package ops

// ServiceGroup ops 模块服务聚合入口。
// 路由/API 通过 service.ServiceGroupApp.OpsServiceGroup.<Xxx>Service 访问。
type ServiceGroup struct {
	AssetService
	CredentialService
	SSHService
	TerminalService
	TicketService
}
