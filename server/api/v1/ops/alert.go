package ops

import (
	"github.com/flipped-aurora/gin-vue-admin/server/model/common/response"
	opsReq "github.com/flipped-aurora/gin-vue-admin/server/model/ops/request"
	"github.com/flipped-aurora/gin-vue-admin/server/utils"
	"github.com/gin-gonic/gin"
)

type AlertApi struct{}

// GetAlertList
// @Tags      Ops
// @Summary   分页获取告警
// @Security  ApiKeyAuth
// @Param     data  body      opsReq.AlertSearch                                                              true  "分页与筛选"
// @Success   200   {object}  response.Response{data=response.PageResult{list=[]ops.Alert},msg=string}         "获取成功"
// @Router    /ops/getAlertList [post]
func (a *AlertApi) GetAlertList(c *gin.Context) {
	var info opsReq.AlertSearch
	if err := c.ShouldBindJSON(&info); err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	list, total, err := alertService.GetList(c.Request.Context(), info)
	if err != nil {
		response.FailWithMessage("获取失败", c)
		return
	}
	response.OkWithDetailed(response.PageResult{
		List: list, Total: total, Page: info.Page, PageSize: info.PageSize,
	}, "获取成功", c)
}

// HandleAlert
// @Tags      Ops
// @Summary   处理告警(标记已处理/忽略)
// @Security  ApiKeyAuth
// @Param     data  body      opsReq.HandleAlertReq                      true  "处理参数"
// @Success   200   {object}  response.Response{msg=string}              "处理成功"
// @Router    /ops/handleAlert [post]
func (a *AlertApi) HandleAlert(c *gin.Context) {
	var req opsReq.HandleAlertReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	handler := utils.GetUserName(c)
	if err := alertService.Handle(c.Request.Context(), req, handler); err != nil {
		response.FailWithMessage("处理失败: "+err.Error(), c)
		return
	}
	response.OkWithMessage("处理成功", c)
}
