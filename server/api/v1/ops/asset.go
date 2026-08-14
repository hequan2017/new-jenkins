package ops

import (
	"github.com/flipped-aurora/gin-vue-admin/server/model/common/response"
	commonReq "github.com/flipped-aurora/gin-vue-admin/server/model/common/request"
	"github.com/flipped-aurora/gin-vue-admin/server/model/ops"
	opsReq "github.com/flipped-aurora/gin-vue-admin/server/model/ops/request"
	"github.com/flipped-aurora/gin-vue-admin/server/utils/logger"
	"github.com/gin-gonic/gin"
)

type AssetApi struct{}

// CreateAsset
// @Tags      Ops
// @Summary   创建资产
// @Security  ApiKeyAuth
// @accept    application/json
// @Produce   application/json
// @Param     data  body      ops.Asset                                 true  "资产定义"
// @Success   200   {object}  response.Response{msg=string}             "创建成功"
// @Router    /ops/createAsset [post]
func (a *AssetApi) CreateAsset(c *gin.Context) {
	var asset ops.Asset
	if err := c.ShouldBindJSON(&asset); err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	if err := assetService.CreateAsset(c.Request.Context(), &asset); err != nil {
		logger.WithCtx(c.Request.Context()).Mod("ops").Err(err).Error("创建资产失败!")
		response.FailWithMessage("创建失败: "+err.Error(), c)
		return
	}
	response.OkWithMessage("创建成功", c)
}

// UpdateAsset
// @Tags      Ops
// @Summary   更新资产
// @Security  ApiKeyAuth
// @accept    application/json
// @Produce   application/json
// @Param     data  body      ops.Asset                                 true  "资产定义(含ID)"
// @Success   200   {object}  response.Response{msg=string}             "更新成功"
// @Router    /ops/updateAsset [put]
func (a *AssetApi) UpdateAsset(c *gin.Context) {
	var asset ops.Asset
	if err := c.ShouldBindJSON(&asset); err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	if err := assetService.UpdateAsset(c.Request.Context(), &asset); err != nil {
		logger.WithCtx(c.Request.Context()).Mod("ops").Err(err).Error("更新资产失败!")
		response.FailWithMessage("更新失败: "+err.Error(), c)
		return
	}
	response.OkWithMessage("更新成功", c)
}

// DeleteAsset
// @Tags      Ops
// @Summary   删除资产
// @Security  ApiKeyAuth
// @accept    application/json
// @Produce   application/json
// @Param     data  body      request.GetById                           true  "资产ID"
// @Success   200   {object}  response.Response{msg=string}             "删除成功"
// @Router    /ops/deleteAsset [delete]
func (a *AssetApi) DeleteAsset(c *gin.Context) {
	var req commonReq.GetById
	if err := c.ShouldBindJSON(&req); err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	if err := assetService.DeleteAsset(c.Request.Context(), req.Uint()); err != nil {
		logger.WithCtx(c.Request.Context()).Mod("ops").Err(err).Error("删除资产失败!")
		response.FailWithMessage("删除失败", c)
		return
	}
	response.OkWithMessage("删除成功", c)
}

// GetAssetList
// @Tags      Ops
// @Summary   分页获取资产列表
// @Security  ApiKeyAuth
// @accept    application/json
// @Produce   application/json
// @Param     data  body      opsReq.AssetSearch                                                  true  "分页与筛选"
// @Success   200   {object}  response.Response{data=response.PageResult{list=[]ops.Asset},msg=string}  "获取成功"
// @Router    /ops/getAssetList [post]
func (a *AssetApi) GetAssetList(c *gin.Context) {
	var info opsReq.AssetSearch
	if err := c.ShouldBindJSON(&info); err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	list, total, err := assetService.GetAssetList(c.Request.Context(), info)
	if err != nil {
		logger.WithCtx(c.Request.Context()).Mod("ops").Err(err).Error("获取资产列表失败!")
		response.FailWithMessage("获取失败", c)
		return
	}
	response.OkWithDetailed(response.PageResult{
		List: list, Total: total, Page: info.Page, PageSize: info.PageSize,
	}, "获取成功", c)
}
