package workflow

import (
	"github.com/flipped-aurora/gin-vue-admin/server/global"
	"gorm.io/datatypes"
)

// 流水线触发方式
const (
	TriggerManual  = "manual"  // 手动触发
	TriggerSchedule = "schedule" // 定时触发(预留,本期不实现调度器)
	TriggerWebhook = "webhook" // Webhook 触发(预留)
)

// Pipeline 流水线定义(对应 Jenkins 的 Job / Pipeline 定义)
// 一个 Pipeline 由多个有序 Stage 组成, 每个 Stage 由多个有序 Step 组成。
// 定义(本表 + Stage + Step)与运行(Build 系列)分离, 类似 Jenkins Job 与 Build 的关系。
type Pipeline struct {
	global.GVA_MODEL
	Name            string         `json:"name" form:"name" gorm:"index;column:name;comment:流水线名称"`
	Description     string         `json:"description" form:"description" gorm:"column:description;comment:流水线说明"`
	TriggerType     string         `json:"triggerType" form:"triggerType" gorm:"column:trigger_type;comment:触发方式 manual|schedule|webhook"`
	Spec            string         `json:"spec" form:"spec" gorm:"column:spec;comment:cron表达式(schedule触发用,支持@daily等描述符)"`
	WithSeconds     bool           `json:"withSeconds" form:"withSeconds" gorm:"column:with_seconds;comment:cron表达式是否含秒位"`
	WebhookSecret   string         `json:"webhookSecret" form:"webhookSecret" gorm:"column:webhook_secret;comment:webhook触发密钥(triggerType=webhook时生成)"`
	ParamSchema     datatypes.JSON `json:"paramSchema" form:"paramSchema" gorm:"column:param_schema;comment:参数定义(JSON 数组)" swaggertype:"object"`
	Enabled         bool           `json:"enabled" form:"enabled" gorm:"column:enabled;comment:是否启用;default:true"`
	// Stages 在创建/更新时级联写入, 查询详情时 Preload 填充
	// form:"-" : 关联对象不参与 query 绑定(参考 backend-layer-rules 的关联对象约束)
	Stages []PipelineStage `json:"stages" form:"-" gorm:"foreignKey:PipelineID;references:ID;constraint:OnDelete:CASCADE"`
}

func (Pipeline) TableName() string {
	return "wf_pipelines"
}

// ParamField 流水线参数定义(序列化后存入 Pipeline.ParamSchema)
// 类型支持 string / number / bool, 前端据此渲染表单。
type ParamField struct {
	Name     string `json:"name" form:"name"`         // 参数键名
	Label    string `json:"label" form:"label"`       // 展示名
	Type     string `json:"type" form:"type"`         // string | number | bool
	Required bool   `json:"required" form:"required"` // 是否必填
	Default  string `json:"default" form:"default"`   // 默认值(字符串形式)
}

// ParamValue 触发构建时的实际入参(序列化后存入 PipelineBuild.Params)
type ParamValue struct {
	Name  string `json:"name" form:"name"`
	Value string `json:"value" form:"value"`
}
