package ops

import (
	"fmt"

	"github.com/flipped-aurora/gin-vue-admin/server/model/common/response"
	opsReq "github.com/flipped-aurora/gin-vue-admin/server/model/ops/request"
	opsmodel "github.com/flipped-aurora/gin-vue-admin/server/model/ops"
	"github.com/flipped-aurora/gin-vue-admin/server/utils"
	"github.com/flipped-aurora/gin-vue-admin/server/utils/logger"
	"github.com/gin-gonic/gin"
)

type FileApi struct{}

// ListDir
// @Tags      Ops
// @Summary   列出远程目录
// @Security  ApiKeyAuth
// @accept    application/json
// @Produce   application/json
// @Param     data  body      opsReq.TestConnectionReq                    true  "资产/凭据"
// @Param     dir   query     string                                      false "目录路径"
// @Success   200   {object}  response.Response{data=object,msg=string}   "获取成功"
// @Router    /ops/listDir [post]
func (a *FileApi) ListDir(c *gin.Context) {
	var req opsReq.TestConnectionReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	dir := c.Query("dir")
	if dir == "" {
		dir = "/"
	}
	entries, err := fileService.ListDir(c.Request.Context(), req.AssetID, req.CredentialID, dir)
	if err != nil {
		logger.WithCtx(c.Request.Context()).Mod("ops").Err(err).Warn("列目录失败")
		response.FailWithMessage("列目录失败: "+err.Error(), c)
		return
	}
	response.OkWithDetailed(gin.H{"dir": dir, "entries": entries}, "获取成功", c)
}

// ReadFile
// @Tags      Ops
// @Summary   读取远程文件内容
// @Security  ApiKeyAuth
// @accept    application/json
// @Produce   application/json
// @Param     data  body      object                                      true  "{assetId,credentialId,path}"
// @Success   200   {object}  response.Response{data=object,msg=string}   "获取成功"
// @Router    /ops/readFile [post]
func (a *FileApi) ReadFile(c *gin.Context) {
	var req struct {
		AssetID      uint   `json:"assetId"`
		CredentialID uint   `json:"credentialId"`
		Path         string `json:"path"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	content, err := fileService.ReadFile(c.Request.Context(), req.AssetID, req.CredentialID, req.Path)
	if err != nil {
		response.FailWithMessage("读取失败: "+err.Error(), c)
		return
	}
	response.OkWithDetailed(gin.H{"path": req.Path, "content": content}, "获取成功", c)
}

// WriteFile
// @Tags      Ops
// @Summary   写入/覆盖远程文件
// @Security  ApiKeyAuth
// @accept    application/json
// @Produce   application/json
// @Param     data  body      object                                      true  "{assetId,credentialId,path,content}"
// @Success   200   {object}  response.Response{msg=string}              "写入成功"
// @Router    /ops/writeFile [post]
func (a *FileApi) WriteFile(c *gin.Context) {
	var req struct {
		AssetID      uint   `json:"assetId"`
		CredentialID uint   `json:"credentialId"`
		Path         string `json:"path"`
		Content      string `json:"content"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	if err := fileService.WriteFile(c.Request.Context(), req.AssetID, req.CredentialID, req.Path, req.Content); err != nil {
		response.FailWithMessage("写入失败: "+err.Error(), c)
		return
	}
	uid := utils.GetUserID(c)
	uname := utils.GetUserName(c)
	_ = auditService.Record(c.Request.Context(), uid, uname, opsmodel.AuditActionFileOp,
		fmt.Sprintf("asset=%d %s", req.AssetID, req.Path), c.ClientIP(), "success", "write")
	response.OkWithMessage("写入成功", c)
}

// RemoveFile
// @Tags      Ops
// @Summary   删除远程文件或空目录
// @Security  ApiKeyAuth
// @accept    application/json
// @Produce   application/json
// @Param     data  body      object                                      true  "{assetId,credentialId,path,isDir}"
// @Success   200   {object}  response.Response{msg=string}              "删除成功"
// @Router    /ops/removeFile [post]
func (a *FileApi) RemoveFile(c *gin.Context) {
	var req struct {
		AssetID      uint   `json:"assetId"`
		CredentialID uint   `json:"credentialId"`
		Path         string `json:"path"`
		IsDir        bool   `json:"isDir"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	if err := fileService.Remove(c.Request.Context(), req.AssetID, req.CredentialID, req.Path, req.IsDir); err != nil {
		response.FailWithMessage("删除失败: "+err.Error(), c)
		return
	}
	uid := utils.GetUserID(c)
	uname := utils.GetUserName(c)
	_ = auditService.Record(c.Request.Context(), uid, uname, opsmodel.AuditActionFileOp,
		fmt.Sprintf("asset=%d %s", req.AssetID, req.Path), c.ClientIP(), "success", "remove")
	response.OkWithMessage("删除成功", c)
}

// RenameFile
// @Tags      Ops
// @Summary   重命名/移动远程文件
// @Security  ApiKeyAuth
// @accept    application/json
// @Produce   application/json
// @Param     data  body      object                                      true  "{assetId,credentialId,oldPath,newPath}"
// @Success   200   {object}  response.Response{msg=string}              "操作成功"
// @Router    /ops/renameFile [post]
func (a *FileApi) RenameFile(c *gin.Context) {
	var req struct {
		AssetID      uint   `json:"assetId"`
		CredentialID uint   `json:"credentialId"`
		OldPath      string `json:"oldPath"`
		NewPath      string `json:"newPath"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	if err := fileService.Rename(c.Request.Context(), req.AssetID, req.CredentialID, req.OldPath, req.NewPath); err != nil {
		response.FailWithMessage("操作失败: "+err.Error(), c)
		return
	}
	response.OkWithMessage("操作成功", c)
}

// Mkdir
// @Tags      Ops
// @Summary   新建远程目录
// @Security  ApiKeyAuth
// @accept    application/json
// @Produce   application/json
// @Param     data  body      object                                      true  "{assetId,credentialId,dir}"
// @Success   200   {object}  response.Response{msg=string}              "创建成功"
// @Router    /ops/mkdir [post]
func (a *FileApi) Mkdir(c *gin.Context) {
	var req struct {
		AssetID      uint   `json:"assetId"`
		CredentialID uint   `json:"credentialId"`
		Dir          string `json:"dir"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	if err := fileService.Mkdir(c.Request.Context(), req.AssetID, req.CredentialID, req.Dir); err != nil {
		response.FailWithMessage("创建失败: "+err.Error(), c)
		return
	}
	response.OkWithMessage("创建成功", c)
}
