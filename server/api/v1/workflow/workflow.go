package workflow

import (
	"github.com/flipped-aurora/gin-vue-admin/server/model/common/response"
	commonReq "github.com/flipped-aurora/gin-vue-admin/server/model/common/request"
	"github.com/flipped-aurora/gin-vue-admin/server/model/workflow"
	workflowReq "github.com/flipped-aurora/gin-vue-admin/server/model/workflow/request"
	"github.com/flipped-aurora/gin-vue-admin/server/utils"
	"github.com/flipped-aurora/gin-vue-admin/server/utils/logger"
	"github.com/flipped-aurora/gin-vue-admin/server/utils/sse"
	"github.com/gin-gonic/gin"
)

type WorkflowApi struct{}

// ============================== 流水线定义 ==============================

// CreatePipeline
// @Tags      Workflow
// @Summary   创建流水线(含阶段/步骤树)
// @Security  ApiKeyAuth
// @accept    application/json
// @Produce   application/json
// @Param     data  body      workflow.Pipeline                                          true  "流水线定义"
// @Success   200   {object}  response.Response{msg=string}                              "创建成功"
// @Router    /workflow/createPipeline [post]
func (a *WorkflowApi) CreatePipeline(c *gin.Context) {
	var p workflow.Pipeline
	if err := c.ShouldBindJSON(&p); err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	if err := pipelineService.CreatePipeline(c.Request.Context(), &p); err != nil {
		logger.WithCtx(c.Request.Context()).Mod("workflow").Err(err).Error("创建流水线失败!")
		response.FailWithMessage("创建失败: "+err.Error(), c)
		return
	}
	response.OkWithMessage("创建成功", c)
}

// UpdatePipeline
// @Tags      Workflow
// @Summary   更新流水线(全量覆盖阶段/步骤)
// @Security  ApiKeyAuth
// @accept    application/json
// @Produce   application/json
// @Param     data  body      workflow.Pipeline                                          true  "流水线定义(含ID)"
// @Success   200   {object}  response.Response{msg=string}                              "更新成功"
// @Router    /workflow/updatePipeline [put]
func (a *WorkflowApi) UpdatePipeline(c *gin.Context) {
	var p workflow.Pipeline
	if err := c.ShouldBindJSON(&p); err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	if err := pipelineService.UpdatePipeline(c.Request.Context(), &p); err != nil {
		logger.WithCtx(c.Request.Context()).Mod("workflow").Err(err).Error("更新流水线失败!")
		response.FailWithMessage("更新失败: "+err.Error(), c)
		return
	}
	response.OkWithMessage("更新成功", c)
}

// DeletePipeline
// @Tags      Workflow
// @Summary   删除流水线
// @Security  ApiKeyAuth
// @accept    application/json
// @Produce   application/json
// @Param     data  body      request.GetById                                            true  "流水线ID"
// @Success   200   {object}  response.Response{msg=string}                              "删除成功"
// @Router    /workflow/deletePipeline [delete]
func (a *WorkflowApi) DeletePipeline(c *gin.Context) {
	var req commonReq.GetById
	if err := c.ShouldBindJSON(&req); err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	if err := pipelineService.DeletePipeline(c.Request.Context(), req.Uint()); err != nil {
		logger.WithCtx(c.Request.Context()).Mod("workflow").Err(err).Error("删除流水线失败!")
		response.FailWithMessage("删除失败", c)
		return
	}
	response.OkWithMessage("删除成功", c)
}

// GetPipelineList
// @Tags      Workflow
// @Summary   分页获取流水线列表
// @Security  ApiKeyAuth
// @accept    application/json
// @Produce   application/json
// @Param     data  body      workflowReq.PipelineSearch                                                              true  "分页与筛选"
// @Success   200   {object}  response.Response{data=response.PageResult{list=[]workflow.Pipeline},msg=string}        "获取成功"
// @Router    /workflow/getPipelineList [post]
func (a *WorkflowApi) GetPipelineList(c *gin.Context) {
	var info workflowReq.PipelineSearch
	if err := c.ShouldBindJSON(&info); err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	list, total, err := pipelineService.GetPipelineList(c.Request.Context(), info)
	if err != nil {
		logger.WithCtx(c.Request.Context()).Mod("workflow").Err(err).Error("获取流水线列表失败!")
		response.FailWithMessage("获取失败", c)
		return
	}
	response.OkWithDetailed(response.PageResult{
		List:     list,
		Total:    total,
		Page:     info.Page,
		PageSize: info.PageSize,
	}, "获取成功", c)
}

// FindPipeline
// @Tags      Workflow
// @Summary   获取流水线详情(含阶段/步骤树)
// @Security  ApiKeyAuth
// @accept    application/json
// @Produce   application/json
// @Param     data  query     request.GetById                                           true  "流水线ID"
// @Success   200   {object}  response.Response{data=workflow.Pipeline,msg=string}      "获取成功"
// @Router    /workflow/findPipeline [get]
func (a *WorkflowApi) FindPipeline(c *gin.Context) {
	var req commonReq.GetById
	if err := c.ShouldBindQuery(&req); err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	p, err := pipelineService.FindPipeline(c.Request.Context(), req.Uint())
	if err != nil {
		logger.WithCtx(c.Request.Context()).Mod("workflow").Err(err).Error("获取流水线详情失败!")
		response.FailWithMessage("获取失败: "+err.Error(), c)
		return
	}
	response.OkWithDetailed(p, "获取成功", c)
}

// ============================== 构建 ==============================

// TriggerBuild
// @Tags      Workflow
// @Summary   触发流水线构建
// @Security  ApiKeyAuth
// @accept    application/json
// @Produce   application/json
// @Param     data  body      workflowReq.TriggerBuildReq                               true  "触发参数"
// @Success   200   {object}  response.Response{data=object,msg=string}                 "已触发, 返回 buildId"
// @Router    /workflow/triggerBuild [post]
func (a *WorkflowApi) TriggerBuild(c *gin.Context) {
	var req workflowReq.TriggerBuildReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	uid := utils.GetUserID(c)
	buildID, err := buildService.TriggerBuild(c.Request.Context(), req.PipelineID, req.Params, workflow.TriggerManual, uid)
	if err != nil {
		logger.WithCtx(c.Request.Context()).Mod("workflow").Err(err).Error("触发构建失败!")
		response.FailWithMessage("触发失败: "+err.Error(), c)
		return
	}
	response.OkWithDetailed(gin.H{"buildId": buildID}, "已触发", c)
}

// CancelBuild
// @Tags      Workflow
// @Summary   取消构建
// @Security  ApiKeyAuth
// @accept    application/json
// @Produce   application/json
// @Param     data  body      request.GetById                                           true  "构建ID"
// @Success   200   {object}  response.Response{msg=string}                             "取消成功"
// @Router    /workflow/cancelBuild [post]
func (a *WorkflowApi) CancelBuild(c *gin.Context) {
	var req commonReq.GetById
	if err := c.ShouldBindJSON(&req); err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	if err := buildService.CancelBuild(c.Request.Context(), req.Uint()); err != nil {
		logger.WithCtx(c.Request.Context()).Mod("workflow").Err(err).Error("取消构建失败!")
		response.FailWithMessage("取消失败: "+err.Error(), c)
		return
	}
	response.OkWithMessage("取消成功", c)
}

// ApproveStage
// @Tags      Workflow
// @Summary   审批 gate: 通过或拒绝
// @Security  ApiKeyAuth
// @accept    application/json
// @Produce   application/json
// @Param     data  body      workflowReq.ApproveStageReq                               true  "审批参数"
// @Success   200   {object}  response.Response{msg=string}                             "审批已提交"
// @Router    /workflow/approveStage [post]
func (a *WorkflowApi) ApproveStage(c *gin.Context) {
	var req workflowReq.ApproveStageReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	if err := buildService.ApproveStage(c.Request.Context(), req); err != nil {
		logger.WithCtx(c.Request.Context()).Mod("workflow").Err(err).Error("审批提交失败!")
		response.FailWithMessage("审批失败: "+err.Error(), c)
		return
	}
	response.OkWithMessage("审批已提交", c)
}

// GetBuildList
// @Tags      Workflow
// @Summary   分页获取构建历史
// @Security  ApiKeyAuth
// @accept    application/json
// @Produce   application/json
// @Param     data  body      workflowReq.PipelineBuildSearch                                                        true  "分页与筛选"
// @Success   200   {object}  response.Response{data=response.PageResult{list=[]workflow.PipelineBuild},msg=string}   "获取成功"
// @Router    /workflow/getBuildList [post]
func (a *WorkflowApi) GetBuildList(c *gin.Context) {
	var info workflowReq.PipelineBuildSearch
	if err := c.ShouldBindJSON(&info); err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	list, total, err := buildService.GetBuildList(c.Request.Context(), info)
	if err != nil {
		logger.WithCtx(c.Request.Context()).Mod("workflow").Err(err).Error("获取构建列表失败!")
		response.FailWithMessage("获取失败", c)
		return
	}
	response.OkWithDetailed(response.PageResult{
		List:     list,
		Total:    total,
		Page:     info.Page,
		PageSize: info.PageSize,
	}, "获取成功", c)
}

// GetBuildDetail
// @Tags      Workflow
// @Summary   获取构建详情(含阶段/步骤运行视图)
// @Security  ApiKeyAuth
// @accept    application/json
// @Produce   application/json
// @Param     data  query     request.GetById                                           true  "构建ID"
// @Success   200   {object}  response.Response{data=workflowRes.BuildDetail,msg=string} "获取成功"
// @Router    /workflow/getBuildDetail [get]
func (a *WorkflowApi) GetBuildDetail(c *gin.Context) {
	var req commonReq.GetById
	if err := c.ShouldBindQuery(&req); err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	detail, err := buildService.GetBuildDetail(c.Request.Context(), req.Uint())
	if err != nil {
		logger.WithCtx(c.Request.Context()).Mod("workflow").Err(err).Error("获取构建详情失败!")
		response.FailWithMessage("获取失败", c)
		return
	}
	response.OkWithDetailed(detail, "获取成功", c)
}

// GetBuildLog
// @Tags      Workflow
// @Summary   分页获取构建日志(按 step 维度)
// @Security  ApiKeyAuth
// @accept    application/json
// @Produce   application/json
// @Param     data  query     workflowReq.BuildLogSearch                                                              true  "分页与筛选"
// @Success   200   {object}  response.Response{data=response.PageResult{list=[]workflow.PipelineBuildLog},msg=string}  "获取成功"
// @Router    /workflow/getBuildLog [get]
func (a *WorkflowApi) GetBuildLog(c *gin.Context) {
	var info workflowReq.BuildLogSearch
	if err := c.ShouldBindQuery(&info); err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	list, total, err := buildService.GetBuildLog(c.Request.Context(), info)
	if err != nil {
		logger.WithCtx(c.Request.Context()).Mod("workflow").Err(err).Error("获取构建日志失败!")
		response.FailWithMessage("获取失败", c)
		return
	}
	response.OkWithDetailed(response.PageResult{
		List:     list,
		Total:    total,
		Page:     info.Page,
		PageSize: info.PageSize,
	}, "获取成功", c)
}

// BuildStream 构建实时事件 SSE 订阅(状态变化 + 日志增量)
// 注意: 本路由绝不能套 TimeoutMiddleware(见 utils/sse/hub.go Stream 注释);
// 订阅按当前登录用户维度, 引擎把事件投递给触发者。
// @Tags      Workflow
// @Summary   订阅构建实时事件(SSE)
// @Security  ApiKeyAuth
// @Produce   application/json
// @Success   200  {object}  response.Response{msg=string}  "SSE 流"
// @Router    /workflow/buildStream [get]
func (a *WorkflowApi) BuildStream(c *gin.Context) {
	userID := utils.GetUserID(c)
	if userID == 0 {
		response.FailWithMessage("未登录", c)
		return
	}
	sse.Default().Stream(c, userID, 0)
}
