package ops

import (
	"github.com/flipped-aurora/gin-vue-admin/server/model/common/response"
	commonReq "github.com/flipped-aurora/gin-vue-admin/server/model/common/request"
	opsReq "github.com/flipped-aurora/gin-vue-admin/server/model/ops/request"
	"github.com/flipped-aurora/gin-vue-admin/server/utils"
	"github.com/flipped-aurora/gin-vue-admin/server/utils/logger"
	"github.com/gin-gonic/gin"
)

type TicketApi struct{}

// CreateTicket
// @Tags      Ops
// @Summary   创建发版工单
// @Security  ApiKeyAuth
// @accept    application/json
// @Produce   application/json
// @Param     data  body      opsReq.TicketInput                       true  "工单内容"
// @Success   200   {object}  response.Response{data=object,msg=string} "创建成功, 返回工单ID"
// @Router    /ops/createTicket [post]
func (a *TicketApi) CreateTicket(c *gin.Context) {
	var in opsReq.TicketInput
	if err := c.ShouldBindJSON(&in); err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	uid := utils.GetUserID(c)
	t, err := ticketService.Create(c.Request.Context(), in, uid)
	if err != nil {
		logger.WithCtx(c.Request.Context()).Mod("ops").Err(err).Error("创建工单失败!")
		response.FailWithMessage("创建失败: "+err.Error(), c)
		return
	}
	response.OkWithDetailed(gin.H{"id": t.ID}, "创建成功", c)
}

// DeleteTicket
// @Tags      Ops
// @Summary   删除工单
// @Security  ApiKeyAuth
// @accept    application/json
// @Produce   application/json
// @Param     data  body      request.GetById                           true  "工单ID"
// @Success   200   {object}  response.Response{msg=string}             "删除成功"
// @Router    /ops/deleteTicket [delete]
func (a *TicketApi) DeleteTicket(c *gin.Context) {
	var req commonReq.GetById
	if err := c.ShouldBindJSON(&req); err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	if err := ticketService.Delete(c.Request.Context(), req.Uint()); err != nil {
		logger.WithCtx(c.Request.Context()).Mod("ops").Err(err).Error("删除工单失败!")
		response.FailWithMessage("删除失败", c)
		return
	}
	response.OkWithMessage("删除成功", c)
}

// GetTicketList
// @Tags      Ops
// @Summary   分页获取工单列表(传 applicantId 仅看本人)
// @Security  ApiKeyAuth
// @accept    application/json
// @Produce   application/json
// @Param     data  body      opsReq.TicketSearch                                                  true  "分页与筛选"
// @Success   200   {object}  response.Response{data=response.PageResult{list=[]ops.Ticket},msg=string}  "获取成功"
// @Router    /ops/getTicketList [post]
func (a *TicketApi) GetTicketList(c *gin.Context) {
	var info opsReq.TicketSearch
	if err := c.ShouldBindJSON(&info); err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	list, total, err := ticketService.GetList(c.Request.Context(), info)
	if err != nil {
		logger.WithCtx(c.Request.Context()).Mod("ops").Err(err).Error("获取工单列表失败!")
		response.FailWithMessage("获取失败", c)
		return
	}
	response.OkWithDetailed(response.PageResult{
		List: list, Total: total, Page: info.Page, PageSize: info.PageSize,
	}, "获取成功", c)
}

// FindTicket
// @Tags      Ops
// @Summary   获取工单详情
// @Security  ApiKeyAuth
// @accept    application/json
// @Produce   application/json
// @Param     data  query     request.GetById                           true  "工单ID"
// @Success   200   {object}  response.Response{data=ops.Ticket,msg=string}  "获取成功"
// @Router    /ops/findTicket [get]
func (a *TicketApi) FindTicket(c *gin.Context) {
	var req commonReq.GetById
	if err := c.ShouldBindQuery(&req); err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	t, err := ticketService.GetByID(c.Request.Context(), req.Uint())
	if err != nil {
		response.FailWithMessage("获取失败: "+err.Error(), c)
		return
	}
	response.OkWithDetailed(t, "获取成功", c)
}

// ApproveTicket
// @Tags      Ops
// @Summary   审批工单(通过则触发流水线构建)
// @Security  ApiKeyAuth
// @accept    application/json
// @Produce   application/json
// @Param     data  body      opsReq.ApproveTicketReq                  true  "审批参数"
// @Success   200   {object}  response.Response{data=object,msg=string} "审批完成"
// @Router    /ops/approveTicket [post]
func (a *TicketApi) ApproveTicket(c *gin.Context) {
	var req opsReq.ApproveTicketReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	uid := utils.GetUserID(c)
	t, err := ticketService.Approve(c.Request.Context(), req, uid)
	if err != nil {
		logger.WithCtx(c.Request.Context()).Mod("ops").Err(err).Warn("工单审批失败")
		response.FailWithMessage("审批失败: "+err.Error(), c)
		return
	}
	response.OkWithDetailed(gin.H{"id": t.ID, "status": t.Status, "buildId": t.BuildID}, "审批完成", c)
}

// CancelTicket
// @Tags      Ops
// @Summary   取消工单(仅待审批状态可取消)
// @Security  ApiKeyAuth
// @accept    application/json
// @Produce   application/json
// @Param     data  body      request.GetById                           true  "工单ID"
// @Success   200   {object}  response.Response{msg=string}             "取消成功"
// @Router    /ops/cancelTicket [post]
func (a *TicketApi) CancelTicket(c *gin.Context) {
	var req commonReq.GetById
	if err := c.ShouldBindJSON(&req); err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	if err := ticketService.Cancel(c.Request.Context(), req.Uint()); err != nil {
		response.FailWithMessage("取消失败: "+err.Error(), c)
		return
	}
	response.OkWithMessage("取消成功", c)
}
