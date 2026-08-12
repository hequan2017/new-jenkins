package workflow

import (
	"errors"
	"fmt"
	"io"
	"strconv"

	"github.com/flipped-aurora/gin-vue-admin/server/model/common/response"
	"github.com/flipped-aurora/gin-vue-admin/server/model/workflow"
	"github.com/flipped-aurora/gin-vue-admin/server/utils/logger"
	"github.com/gin-gonic/gin"
)

// WebhookTrigger 公开 webhook 触发入口(免登录, 靠 X-Webhook-Secret 头校验)
// 由 router/workflow 的 InitWebhookRouter 挂到 PublicGroup。
// 请求体 JSON 解析为 map, 键名匹配 ParamSchema 转为 []ParamValue。
//
// @Tags      Workflow
// @Summary   Webhook 触发流水线(公开, 需 X-Webhook-Secret 头)
// @accept    application/json
// @Produce   application/json
// @Param     id     path   int    true  "流水线ID"
// @Param     X-Webhook-Secret  header  string  true  "webhook 密钥"
// @Success   200   {object}  response.Response{data=object,msg=string}  "已触发, 返回 buildId"
// @Router    /webhook/trigger/{id} [post]
func (a *WorkflowApi) WebhookTrigger(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 64)
	if err != nil {
		response.FailWithMessage("流水线ID非法", c)
		return
	}
	secret := c.GetHeader("X-Webhook-Secret")

	p, err := pipelineService.AuthenticateWebhook(c.Request.Context(), uint(id), secret)
	if err != nil {
		response.FailWithMessage("鉴权失败", c)
		return
	}

	// 解析请求体为参数(键值对 -> []ParamValue)
	var bodyMap map[string]any
	if err := c.ShouldBindJSON(&bodyMap); err != nil && !errors.Is(err, io.EOF) {
		response.FailWithMessage("请求体必须是 JSON 对象", c)
		return
	}
	params := make([]workflow.ParamValue, 0, len(bodyMap))
	for k, v := range bodyMap {
		switch v.(type) {
		case string, float64, bool:
			params = append(params, workflow.ParamValue{Name: k, Value: fmt.Sprint(v)})
		case nil:
			params = append(params, workflow.ParamValue{Name: k})
		default:
			response.FailWithMessage("Webhook 参数只支持 string、number、bool", c)
			return
		}
	}

	buildID, err := buildService.TriggerBuild(c.Request.Context(), p.ID, params, workflow.TriggerWebhook, 0)
	if err != nil {
		logger.WithCtx(c.Request.Context()).Mod("workflow").Err(err).Error("webhook 触发失败!")
		response.FailWithMessage("触发失败", c)
		return
	}
	response.OkWithDetailed(gin.H{"buildId": buildID}, "已触发", c)
}
