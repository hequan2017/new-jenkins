package ops

import (
	"github.com/flipped-aurora/gin-vue-admin/server/model/common/response"
	commonReq "github.com/flipped-aurora/gin-vue-admin/server/model/common/request"
	"github.com/flipped-aurora/gin-vue-admin/server/model/ops"
	opsReq "github.com/flipped-aurora/gin-vue-admin/server/model/ops/request"
	"github.com/flipped-aurora/gin-vue-admin/server/utils/logger"
	"github.com/gin-gonic/gin"
)

type AssetGroupApi struct{}

// CreateAssetGroup
// @Tags      Ops
// @Summary   创建资产分组
// @Security  ApiKeyAuth
// @accept    application/json
// @Produce   application/json
// @Param     data  body      ops.AssetGroup                             true  "分组定义"
// @Success   200   {object}  response.Response{msg=string}              "创建成功"
// @Router    /ops/createAssetGroup [post]
func (a *AssetGroupApi) CreateAssetGroup(c *gin.Context) {
	var g ops.AssetGroup
	if err := c.ShouldBindJSON(&g); err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	if err := assetGroupService.Create(c.Request.Context(), &g); err != nil {
		logger.WithCtx(c.Request.Context()).Mod("ops").Err(err).Error("创建资产分组失败!")
		response.FailWithMessage("创建失败: "+err.Error(), c)
		return
	}
	response.OkWithMessage("创建成功", c)
}

// UpdateAssetGroup
// @Tags      Ops
// @Summary   更新资产分组
// @Security  ApiKeyAuth
// @Param     data  body      ops.AssetGroup                             true  "分组定义(含ID)"
// @Success   200   {object}  response.Response{msg=string}              "更新成功"
// @Router    /ops/updateAssetGroup [put]
func (a *AssetGroupApi) UpdateAssetGroup(c *gin.Context) {
	var g ops.AssetGroup
	if err := c.ShouldBindJSON(&g); err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	if err := assetGroupService.Update(c.Request.Context(), &g); err != nil {
		response.FailWithMessage("更新失败: "+err.Error(), c)
		return
	}
	response.OkWithMessage("更新成功", c)
}

// DeleteAssetGroup
// @Tags      Ops
// @Summary   删除资产分组
// @Security  ApiKeyAuth
// @Param     data  body      request.GetById                            true  "分组ID"
// @Success   200   {object}  response.Response{msg=string}              "删除成功"
// @Router    /ops/deleteAssetGroup [delete]
func (a *AssetGroupApi) DeleteAssetGroup(c *gin.Context) {
	var req commonReq.GetById
	if err := c.ShouldBindJSON(&req); err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	if err := assetGroupService.Delete(c.Request.Context(), req.Uint()); err != nil {
		response.FailWithMessage("删除失败: "+err.Error(), c)
		return
	}
	response.OkWithMessage("删除成功", c)
}

// GetAssetGroupList
// @Tags      Ops
// @Summary   分页获取资产分组
// @Security  ApiKeyAuth
// @Param     data  body      opsReq.AssetGroupSearch                                                          true  "分页与筛选"
// @Success   200   {object}  response.Response{data=response.PageResult{list=[]ops.AssetGroup},msg=string}      "获取成功"
// @Router    /ops/getAssetGroupList [post]
func (a *AssetGroupApi) GetAssetGroupList(c *gin.Context) {
	var info opsReq.AssetGroupSearch
	if err := c.ShouldBindJSON(&info); err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	list, total, err := assetGroupService.GetList(c.Request.Context(), info)
	if err != nil {
		response.FailWithMessage("获取失败", c)
		return
	}
	response.OkWithDetailed(response.PageResult{
		List: list, Total: total, Page: info.Page, PageSize: info.PageSize,
	}, "获取成功", c)
}

// GetAllAssetGroups
// @Tags      Ops
// @Summary   获取全部资产分组(下拉用)
// @Security  ApiKeyAuth
// @Success   200   {object}  response.Response{data=[]ops.AssetGroup,msg=string}  "获取成功"
// @Router    /ops/getAllAssetGroups [get]
func (a *AssetGroupApi) GetAllAssetGroups(c *gin.Context) {
	list, err := assetGroupService.GetAll(c.Request.Context())
	if err != nil {
		response.FailWithMessage("获取失败", c)
		return
	}
	response.OkWithDetailed(list, "获取成功", c)
}
