package ops

import (
	"os"

	"github.com/flipped-aurora/gin-vue-admin/server/model/common/response"
	commonReq "github.com/flipped-aurora/gin-vue-admin/server/model/common/request"
	opsReq "github.com/flipped-aurora/gin-vue-admin/server/model/ops/request"
	"github.com/flipped-aurora/gin-vue-admin/server/utils/logger"
	"github.com/gin-gonic/gin"
)

type BackupApi struct{}

// CreateBackupTask
// @Tags      Ops
// @Summary   创建备份任务
// @Security  ApiKeyAuth
// @Param     data  body      opsReq.BackupTaskInput                     true  "备份任务定义"
// @Success   200   {object}  response.Response{data=object,msg=string}  "创建成功"
// @Router    /ops/createBackupTask [post]
func (a *BackupApi) CreateBackupTask(c *gin.Context) {
	var in opsReq.BackupTaskInput
	if err := c.ShouldBindJSON(&in); err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	t, err := backupService.Create(c.Request.Context(), in)
	if err != nil {
		logger.WithCtx(c.Request.Context()).Mod("ops").Err(err).Error("创建备份任务失败!")
		response.FailWithMessage("创建失败: "+err.Error(), c)
		return
	}
	response.OkWithDetailed(gin.H{"id": t.ID}, "创建成功", c)
}

// UpdateBackupTask
// @Tags      Ops
// @Summary   更新备份任务
// @Security  ApiKeyAuth
// @Param     data  body      opsReq.BackupTaskInput                     true  "备份任务定义(含ID)"
// @Success   200   {object}  response.Response{msg=string}              "更新成功"
// @Router    /ops/updateBackupTask [put]
func (a *BackupApi) UpdateBackupTask(c *gin.Context) {
	var req struct {
		ID uint `json:"id"`
		opsReq.BackupTaskInput
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	if err := backupService.Update(c.Request.Context(), req.ID, req.BackupTaskInput); err != nil {
		response.FailWithMessage("更新失败: "+err.Error(), c)
		return
	}
	response.OkWithMessage("更新成功", c)
}

// ToggleBackupTask
// @Tags      Ops
// @Summary   启停备份任务
// @Security  ApiKeyAuth
// @Param     data  body      object                                     true  "{id,enabled}"
// @Success   200   {object}  response.Response{msg=string}              "操作成功"
// @Router    /ops/toggleBackupTask [post]
func (a *BackupApi) ToggleBackupTask(c *gin.Context) {
	var req struct {
		ID      uint `json:"id"`
		Enabled bool `json:"enabled"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	if err := backupService.Toggle(c.Request.Context(), req.ID, req.Enabled); err != nil {
		response.FailWithMessage("操作失败: "+err.Error(), c)
		return
	}
	response.OkWithMessage("操作成功", c)
}

// RunBackupTask
// @Tags      Ops
// @Summary   立即执行一次备份
// @Security  ApiKeyAuth
// @Param     data  body      request.GetById                            true  "任务ID"
// @Success   200   {object}  response.Response{msg=string}              "已触发"
// @Router    /ops/runBackupTask [post]
func (a *BackupApi) RunBackupTask(c *gin.Context) {
	var req commonReq.GetById
	if err := c.ShouldBindJSON(&req); err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	if err := backupService.RunNow(c.Request.Context(), req.Uint()); err != nil {
		response.FailWithMessage("触发失败: "+err.Error(), c)
		return
	}
	response.OkWithMessage("已触发", c)
}

// DeleteBackupTask
// @Tags      Ops
// @Summary   删除备份任务
// @Security  ApiKeyAuth
// @Param     data  body      request.GetById                            true  "任务ID"
// @Success   200   {object}  response.Response{msg=string}              "删除成功"
// @Router    /ops/deleteBackupTask [delete]
func (a *BackupApi) DeleteBackupTask(c *gin.Context) {
	var req commonReq.GetById
	if err := c.ShouldBindJSON(&req); err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	if err := backupService.Delete(c.Request.Context(), req.Uint()); err != nil {
		response.FailWithMessage("删除失败", c)
		return
	}
	response.OkWithMessage("删除成功", c)
}

// GetBackupTaskList
// @Tags      Ops
// @Summary   分页获取备份任务
// @Security  ApiKeyAuth
// @Param     data  body      opsReq.BackupTaskSearch                                                       true  "分页与筛选"
// @Success   200   {object}  response.Response{data=response.PageResult{list=[]ops.BackupTask},msg=string}   "获取成功"
// @Router    /ops/getBackupTaskList [post]
func (a *BackupApi) GetBackupTaskList(c *gin.Context) {
	var info opsReq.BackupTaskSearch
	if err := c.ShouldBindJSON(&info); err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	list, total, err := backupService.GetTaskList(c.Request.Context(), info)
	if err != nil {
		response.FailWithMessage("获取失败", c)
		return
	}
	response.OkWithDetailed(response.PageResult{
		List: list, Total: total, Page: info.Page, PageSize: info.PageSize,
	}, "获取成功", c)
}

// GetBackupRecordList
// @Tags      Ops
// @Summary   分页获取备份记录
// @Security  ApiKeyAuth
// @Param     data  body      opsReq.BackupRecordSearch                                                       true  "分页与筛选"
// @Success   200   {object}  response.Response{data=response.PageResult{list=[]ops.BackupRecord},msg=string}   "获取成功"
// @Router    /ops/getBackupRecordList [post]
func (a *BackupApi) GetBackupRecordList(c *gin.Context) {
	var info opsReq.BackupRecordSearch
	if err := c.ShouldBindJSON(&info); err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	list, total, err := backupService.GetRecordList(c.Request.Context(), info)
	if err != nil {
		response.FailWithMessage("获取失败", c)
		return
	}
	response.OkWithDetailed(response.PageResult{
		List: list, Total: total, Page: info.Page, PageSize: info.PageSize,
	}, "获取成功", c)
}

// DownloadBackup
// @Tags      Ops
// @Summary   下载备份归档文件
// @Security  ApiKeyAuth
// @Param     recordId  query  int  true  "备份记录ID"
// @Success   200       {file}  binary  "文件流"
// @Router    /ops/downloadBackup [get]
func (a *BackupApi) DownloadBackup(c *gin.Context) {
	recordID := parseUint(c.Query("recordId"))
	if recordID == 0 {
		response.FailWithMessage("缺少 recordId", c)
		return
	}
	path, err := backupService.ReadArchive(recordID)
	if err != nil {
		response.FailWithMessage("下载失败: "+err.Error(), c)
		return
	}
	f, err := os.Open(path)
	if err != nil {
		response.FailWithMessage("打开文件失败", c)
		return
	}
	defer f.Close()
	info, _ := os.Stat(path)
	c.FileAttachment(path, info.Name())
}
