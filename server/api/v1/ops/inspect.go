package ops

import (
	"github.com/flipped-aurora/gin-vue-admin/server/model/common/response"
	commonReq "github.com/flipped-aurora/gin-vue-admin/server/model/common/request"
	opsReq "github.com/flipped-aurora/gin-vue-admin/server/model/ops/request"
	"github.com/flipped-aurora/gin-vue-admin/server/utils/logger"
	"github.com/gin-gonic/gin"
)

type InspectApi struct{}

// CreateInspectTask
// @Tags      Ops
// @Summary   创建巡检任务
// @Security  ApiKeyAuth
// @accept    application/json
// @Produce   application/json
// @Param     data  body      opsReq.InspectTaskInput                    true  "巡检任务定义"
// @Success   200   {object}  response.Response{data=object,msg=string}  "创建成功, 返回任务ID"
// @Router    /ops/createInspectTask [post]
func (a *InspectApi) CreateInspectTask(c *gin.Context) {
	var in opsReq.InspectTaskInput
	if err := c.ShouldBindJSON(&in); err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	t, err := inspectService.Create(c.Request.Context(), in)
	if err != nil {
		logger.WithCtx(c.Request.Context()).Mod("ops").Err(err).Error("创建巡检任务失败!")
		response.FailWithMessage("创建失败: "+err.Error(), c)
		return
	}
	response.OkWithDetailed(gin.H{"id": t.ID}, "创建成功", c)
}

// UpdateInspectTask
// @Tags      Ops
// @Summary   更新巡检任务
// @Security  ApiKeyAuth
// @accept    application/json
// @Produce   application/json
// @Param     data  body      opsReq.InspectTaskInput                    true  "巡检任务定义(含ID)"
// @Success   200   {object}  response.Response{msg=string}              "更新成功"
// @Router    /ops/updateInspectTask [put]
func (a *InspectApi) UpdateInspectTask(c *gin.Context) {
	var req struct {
		ID uint `json:"id"`
		opsReq.InspectTaskInput
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	if err := inspectService.Update(c.Request.Context(), req.ID, req.InspectTaskInput); err != nil {
		logger.WithCtx(c.Request.Context()).Mod("ops").Err(err).Error("更新巡检任务失败!")
		response.FailWithMessage("更新失败: "+err.Error(), c)
		return
	}
	response.OkWithMessage("更新成功", c)
}

// ToggleInspectTask
// @Tags      Ops
// @Summary   启停巡检任务
// @Security  ApiKeyAuth
// @accept    application/json
// @Produce   application/json
// @Param     data  body      object                                      true  "{id,enabled}"
// @Success   200   {object}  response.Response{msg=string}              "操作成功"
// @Router    /ops/toggleInspectTask [post]
func (a *InspectApi) ToggleInspectTask(c *gin.Context) {
	var req struct {
		ID      uint `json:"id"`
		Enabled bool `json:"enabled"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	if err := inspectService.Toggle(c.Request.Context(), req.ID, req.Enabled); err != nil {
		response.FailWithMessage("操作失败: "+err.Error(), c)
		return
	}
	response.OkWithMessage("操作成功", c)
}

// RunInspectTask
// @Tags      Ops
// @Summary   立即执行一次巡检
// @Security  ApiKeyAuth
// @accept    application/json
// @Produce   application/json
// @Param     data  body      request.GetById                             true  "任务ID"
// @Success   200   {object}  response.Response{msg=string}              "已触发"
// @Router    /ops/runInspectTask [post]
func (a *InspectApi) RunInspectTask(c *gin.Context) {
	var req commonReq.GetById
	if err := c.ShouldBindJSON(&req); err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	if err := inspectService.RunNow(c.Request.Context(), req.Uint()); err != nil {
		response.FailWithMessage("触发失败: "+err.Error(), c)
		return
	}
	response.OkWithMessage("已触发", c)
}

// DeleteInspectTask
// @Tags      Ops
// @Summary   删除巡检任务
// @Security  ApiKeyAuth
// @accept    application/json
// @Produce   application/json
// @Param     data  body      request.GetById                             true  "任务ID"
// @Success   200   {object}  response.Response{msg=string}              "删除成功"
// @Router    /ops/deleteInspectTask [delete]
func (a *InspectApi) DeleteInspectTask(c *gin.Context) {
	var req commonReq.GetById
	if err := c.ShouldBindJSON(&req); err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	if err := inspectService.Delete(c.Request.Context(), req.Uint()); err != nil {
		response.FailWithMessage("删除失败: "+err.Error(), c)
		return
	}
	response.OkWithMessage("删除成功", c)
}

// GetInspectTaskList
// @Tags      Ops
// @Summary   分页获取巡检任务
// @Security  ApiKeyAuth
// @accept    application/json
// @Produce   application/json
// @Param     data  body      opsReq.InspectTaskSearch                                                       true  "分页与筛选"
// @Success   200   {object}  response.Response{data=response.PageResult{list=[]ops.InspectTask},msg=string}   "获取成功"
// @Router    /ops/getInspectTaskList [post]
func (a *InspectApi) GetInspectTaskList(c *gin.Context) {
	var info opsReq.InspectTaskSearch
	if err := c.ShouldBindJSON(&info); err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	list, total, err := inspectService.GetTaskList(c.Request.Context(), info)
	if err != nil {
		response.FailWithMessage("获取失败", c)
		return
	}
	response.OkWithDetailed(response.PageResult{
		List: list, Total: total, Page: info.Page, PageSize: info.PageSize,
	}, "获取成功", c)
}

// GetInspectResultList
// @Tags      Ops
// @Summary   分页获取巡检结果
// @Security  ApiKeyAuth
// @accept    application/json
// @Produce   application/json
// @Param     data  body      opsReq.InspectResultSearch                                                       true  "分页与筛选"
// @Success   200   {object}  response.Response{data=response.PageResult{list=[]ops.InspectResult},msg=string}  "获取成功"
// @Router    /ops/getInspectResultList [post]
func (a *InspectApi) GetInspectResultList(c *gin.Context) {
	var info opsReq.InspectResultSearch
	if err := c.ShouldBindJSON(&info); err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	list, total, err := inspectService.GetResultList(c.Request.Context(), info)
	if err != nil {
		response.FailWithMessage("获取失败", c)
		return
	}
	response.OkWithDetailed(response.PageResult{
		List: list, Total: total, Page: info.Page, PageSize: info.PageSize,
	}, "获取成功", c)
}
