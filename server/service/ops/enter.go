package ops

// ServiceGroup ops 模块服务聚合入口。
type ServiceGroup struct {
	AssetService
	CredentialService
	SSHService
	TerminalService
	TicketService
	AuditService
	InspectService
	DashboardService
	FileService
}
