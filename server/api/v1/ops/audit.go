package ops

import (
	"github.com/flipped-aurora/gin-vue-admin/server/model/common/response"
	opsReq "github.com/flipped-aurora/gin-vue-admin/server/model/ops/request"
	"github.com/flipped-aurora/gin-vue-admin/server/utils/logger"
	"github.com/gin-gonic/gin"
)

type AuditApi struct{}

// GetAuditList
// @Tags      Ops
// @Summary   分页获取运维审计日志
// @Security  ApiKeyAuth
// @accept    application/json
// @Produce   application/json
// @Param     data  body      opsReq.AuditSearch                                                              true  "分页与筛选"
// @Success   200   {object}  response.Response{data=response.PageResult{list=[]ops.AuditRecord},msg=string}  "获取成功"
// @Router    /ops/getAuditList [post]
func (a *AuditApi) GetAuditList(c *gin.Context) {
	var info opsReq.AuditSearch
	if err := c.ShouldBindJSON(&info); err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	list, total, err := auditService.GetList(c.Request.Context(), info)
	if err != nil {
		logger.WithCtx(c.Request.Context()).Mod("ops").Err(err).Error("获取审计列表失败!")
		response.FailWithMessage("获取失败", c)
		return
	}
	response.OkWithDetailed(response.PageResult{
		List: list, Total: total, Page: info.Page, PageSize: info.PageSize,
	}, "获取成功", c)
}
