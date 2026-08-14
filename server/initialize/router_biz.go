package initialize

import (
	"github.com/flipped-aurora/gin-vue-admin/server/router"
	"github.com/gin-gonic/gin"
)

// 占位方法，保证文件可以正确加载，避免go空变量检测报错，请勿删除。
func holder(routers ...*gin.RouterGroup) {
	_ = routers
	_ = router.RouterGroupApp
}

func initBizRouter(routers ...*gin.RouterGroup) {
	privateGroup := routers[0]
	publicGroup := routers[1]

	// 业务模块路由注册
	router.RouterGroupApp.Workflow.InitWorkflowRouter(privateGroup)
	router.RouterGroupApp.Workflow.InitWebhookRouter(publicGroup) // webhook 公开触发入口
	router.RouterGroupApp.Ops.InitOpsRouter(privateGroup)         // 运维模块: 资产/跳板机/工单

	holder(publicGroup, privateGroup)

}
