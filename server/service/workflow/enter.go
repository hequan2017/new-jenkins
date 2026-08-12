package workflow

// ServiceGroup workflow 模块服务聚合入口
// 路由/API 通过 service.ServiceGroupApp.WorkflowServiceGroup 访问。
type ServiceGroup struct {
	PipelineService
	BuildService
}
