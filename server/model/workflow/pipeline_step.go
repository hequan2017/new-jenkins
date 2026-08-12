package workflow

import (
	"github.com/flipped-aurora/gin-vue-admin/server/global"
	"gorm.io/datatypes"
)

// Step 类型
const (
	StepTypeHTTP  = "http"  // HTTP 回调:发请求, 状态码 2xx 视为成功
	StepTypeShell = "shell" // 本地 Shell 命令:GVA 进程内 os/exec 执行, 退出码 0 视为成功
)

// PipelineStep 流水线步骤定义(对应 Jenkins 的 Step / sh / httpRequest)
// Config 按 Type 有不同结构:
//   - http:  { url, method, headers, body, timeoutSec, allowPrivate }
//   - shell: { command, timeoutSec, env }
// 具体解析在 service/workflow/executor 内完成。
type PipelineStep struct {
	global.GVA_MODEL
	StageID uint           `json:"stageId" form:"stageId" gorm:"index;column:stage_id;comment:所属阶段ID"`
	Name    string         `json:"name" form:"name" gorm:"column:name;comment:步骤名称"`
	Type    string         `json:"type" form:"type" gorm:"column:type;comment:步骤类型 http|shell"`
	Config  datatypes.JSON `json:"config" form:"config" gorm:"column:config;comment:步骤配置(按 type 不同结构)" swaggertype:"object"`
	Order   int            `json:"order" form:"order" gorm:"column:sort_order;comment:步骤顺序(升序)"`
}

func (PipelineStep) TableName() string {
	return "wf_pipeline_steps"
}
