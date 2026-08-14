package ops

import (
	"github.com/flipped-aurora/gin-vue-admin/server/model/common/response"
	commonReq "github.com/flipped-aurora/gin-vue-admin/server/model/common/request"
	opsReq "github.com/flipped-aurora/gin-vue-admin/server/model/ops/request"
	"github.com/flipped-aurora/gin-vue-admin/server/utils/logger"
	"github.com/gin-gonic/gin"
)

type CredentialApi struct{}

// CreateCredential
// @Tags      Ops
// @Summary   创建凭据(密码/私钥加密存储)
// @Security  ApiKeyAuth
// @accept    application/json
// @Produce   application/json
// @Param     data  body      opsReq.CredentialInput                    true  "凭据定义(Secret/Passphrase 为明文)"
// @Success   200   {object}  response.Response{data=object,msg=string} "创建成功, 返回凭据ID"
// @Router    /ops/createCredential [post]
func (a *CredentialApi) CreateCredential(c *gin.Context) {
	var in opsReq.CredentialInput
	if err := c.ShouldBindJSON(&in); err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	cred, err := credentialService.CreateCredential(c.Request.Context(), in)
	if err != nil {
		logger.WithCtx(c.Request.Context()).Mod("ops").Err(err).Error("创建凭据失败!")
		response.FailWithMessage("创建失败: "+err.Error(), c)
		return
	}
	response.OkWithDetailed(gin.H{"id": cred.ID}, "创建成功", c)
}

// UpdateCredential
// @Tags      Ops
// @Summary   更新凭据(Secret/Passphrase 留空表示不改)
// @Security  ApiKeyAuth
// @accept    application/json
// @Produce   application/json
// @Param     data  body      opsReq.CredentialInput                    true  "凭据定义(含ID)"
// @Success   200   {object}  response.Response{msg=string}             "更新成功"
// @Router    /ops/updateCredential [put]
func (a *CredentialApi) UpdateCredential(c *gin.Context) {
	var req struct {
		ID uint `json:"id"`
		opsReq.CredentialInput
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	if err := credentialService.UpdateCredential(c.Request.Context(), req.ID, req.CredentialInput); err != nil {
		logger.WithCtx(c.Request.Context()).Mod("ops").Err(err).Error("更新凭据失败!")
		response.FailWithMessage("更新失败: "+err.Error(), c)
		return
	}
	response.OkWithMessage("更新成功", c)
}

// DeleteCredential
// @Tags      Ops
// @Summary   删除凭据
// @Security  ApiKeyAuth
// @accept    application/json
// @Produce   application/json
// @Param     data  body      request.GetById                           true  "凭据ID"
// @Success   200   {object}  response.Response{msg=string}             "删除成功"
// @Router    /ops/deleteCredential [delete]
func (a *CredentialApi) DeleteCredential(c *gin.Context) {
	var req commonReq.GetById
	if err := c.ShouldBindJSON(&req); err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	if err := credentialService.DeleteCredential(c.Request.Context(), req.Uint()); err != nil {
		logger.WithCtx(c.Request.Context()).Mod("ops").Err(err).Error("删除凭据失败!")
		response.FailWithMessage("删除失败", c)
		return
	}
	response.OkWithMessage("删除成功", c)
}

// GetCredentialList
// @Tags      Ops
// @Summary   分页获取凭据列表(不返回 Secret 明文)
// @Security  ApiKeyAuth
// @accept    application/json
// @Produce   application/json
// @Param     data  body      opsReq.CredentialSearch                                            true  "分页与筛选"
// @Success   200   {object}  response.Response{data=response.PageResult{list=[]ops.Credential},msg=string}  "获取成功"
// @Router    /ops/getCredentialList [post]
func (a *CredentialApi) GetCredentialList(c *gin.Context) {
	var info opsReq.CredentialSearch
	if err := c.ShouldBindJSON(&info); err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	list, total, err := credentialService.GetCredentialList(c.Request.Context(), info)
	if err != nil {
		logger.WithCtx(c.Request.Context()).Mod("ops").Err(err).Error("获取凭据列表失败!")
		response.FailWithMessage("获取失败", c)
		return
	}
	response.OkWithDetailed(response.PageResult{
		List: list, Total: total, Page: info.Page, PageSize: info.PageSize,
	}, "获取成功", c)
}
