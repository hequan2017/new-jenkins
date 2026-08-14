package ops

// ServiceGroup ops 模块服务聚合入口。
type ServiceGroup struct {
	AssetService
	AssetGroupService
	CredentialService
	SSHService
	TerminalService
	TicketService
	AuditService
	InspectService
	BackupService
	DashboardService
	FileService
	AlertService
	ScheduleCenterService
}
