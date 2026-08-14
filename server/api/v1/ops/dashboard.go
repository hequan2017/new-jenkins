package ops

import (
	"github.com/flipped-aurora/gin-vue-admin/server/model/common/response"
	"github.com/flipped-aurora/gin-vue-admin/server/utils/logger"
	"github.com/gin-gonic/gin"
)

type DashboardApi struct{}

// GetDashboard
// @Tags      Ops
// @Summary   获取运维大盘聚合数据
// @Security  ApiKeyAuth
// @Produce   application/json
// @Success   200   {object}  response.Response{data=object,msg=string}  "获取成功"
// @Router    /ops/getDashboard [get]
func (a *DashboardApi) GetDashboard(c *gin.Context) {
	data, err := dashboardService.GetDashboard(c.Request.Context())
	if err != nil {
		logger.WithCtx(c.Request.Context()).Mod("ops").Err(err).Error("获取大盘数据失败!")
		response.FailWithMessage("获取失败", c)
		return
	}
	response.OkWithDetailed(data, "获取成功", c)
}
