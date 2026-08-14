package ops

import (
	"github.com/flipped-aurora/gin-vue-admin/server/model/common/response"
	"github.com/gin-gonic/gin"
)

type ScheduleCenterApi struct{}

// GetScheduleList
// @Tags      Ops
// @Summary   获取聚合调度任务列表(workflow/inspect/backup)
// @Security  ApiKeyAuth
// @Param     source  query  string  false  "来源过滤 workflow|inspect|backup"
// @Success   200     {object}  response.Response{data=object,msg=string}  "获取成功"
// @Router    /ops/getScheduleList [get]
func (a *ScheduleCenterApi) GetScheduleList(c *gin.Context) {
	source := c.Query("source")
	list, total, err := scheduleCenterService.GetList(c.Request.Context(), source)
	if err != nil {
		response.FailWithMessage("获取失败", c)
		return
	}
	response.OkWithDetailed(gin.H{"list": list, "total": total}, "获取成功", c)
}
