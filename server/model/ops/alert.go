package ops

import "github.com/flipped-aurora/gin-vue-admin/server/global"

// 告警状态
const (
	AlertStatusActive   = "active"   // 未处理
	AlertStatusResolved = "resolved" // 已处理
	AlertStatusIgnored  = "ignored"  // 已忽略
)

// 告警级别
const (
	AlertLevelInfo     = "info"
	AlertLevelWarning  = "warning"
	AlertLevelCritical = "critical"
)

// 告警来源类型
const (
	AlertSourceInspect = "inspect" // 巡检异常
	AlertSourceTicket  = "ticket"  // 工单状态变更
	AlertSourceBackup  = "backup"  // 备份失败
	AlertSourceManual  = "manual"  // 手动
)

// Alert 告警事件: 由巡检异常/工单变更/备份失败等产生, 统一在告警中心查看与处理。
type Alert struct {
	global.GVA_MODEL
	Title    string `json:"title" form:"title" gorm:"column:title;comment:告警标题"`
	Source   string `json:"source" form:"source" gorm:"index;column:source;comment:来源 inspect|ticket|backup|manual"`
	Level    string `json:"level" form:"level" gorm:"column:level;comment:级别 info|warning|critical;default:warning"`
	RefID    uint   `json:"refId" form:"refId" gorm:"column:ref_id;comment:关联对象ID(如巡检任务ID/工单ID)"`
	RefName  string `json:"refName" form:"refName" gorm:"column:ref_name;comment:关联对象名称"`
	Status   string `json:"status" form:"status" gorm:"index;column:status;comment:状态 active|resolved|ignored;default:active"`
	Detail   string `json:"detail" form:"detail" gorm:"column:detail;type:text;comment:详情"`
	Handler  string `json:"handler" form:"handler" gorm:"column:handler;comment:处理人"`
	Comment  string `json:"comment" form:"comment" gorm:"column:comment;type:text;comment:处理意见"`
}

func (Alert) TableName() string { return "ops_alerts" }
