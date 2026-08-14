package ops

import (
	"github.com/flipped-aurora/gin-vue-admin/server/global"
)

// 巡检结果状态
const (
	InspectStatusOK    = "ok"    // 正常
	InspectStatusAlert = "alert" // 异常(命中关键字或命令失败)
)

// InspectTask 巡检任务: 周期性 SSH 到资产执行检查命令, 命中异常关键字则告警。
// 复用 global.GVA_TIMER 调度, 复用跳板机 SSH 能力执行命令。
type InspectTask struct {
	global.GVA_MODEL
	Name         string `json:"name" form:"name" gorm:"index;column:name;comment:任务名称"`
	AssetID      uint   `json:"assetId" form:"assetId" gorm:"column:asset_id;comment:目标资产ID"`
	CredentialID uint   `json:"credentialId" form:"credentialId" gorm:"column:credential_id;comment:凭据ID"`
	Command      string `json:"command" form:"command" gorm:"column:command;comment:巡检命令"`
	Keyword      string `json:"keyword" form:"keyword" gorm:"column:keyword;comment:异常关键字(逗号分隔)"`
	Spec         string `json:"spec" form:"spec" gorm:"column:spec;comment:cron表达式"`
	WithSeconds  bool   `json:"withSeconds" form:"withSeconds" gorm:"column:with_seconds;comment:cron是否含秒位"`
	Enabled      bool   `json:"enabled" form:"enabled" gorm:"column:enabled;comment:是否启用;default:false"`
	Remark       string `json:"remark" form:"remark" gorm:"column:remark;comment:备注"`
}

func (InspectTask) TableName() string { return "ops_inspect_tasks" }

// InspectResult 巡检结果(每次执行一条)
type InspectResult struct {
	global.GVA_MODEL
	TaskID uint   `json:"taskId" form:"taskId" gorm:"index;column:task_id;comment:任务ID"`
	Status string `json:"status" form:"status" gorm:"index;column:status;comment:结果状态 ok|alert"`
	Output string `json:"output" form:"output" gorm:"column:output;type:text;comment:命令输出(截断)"`
	Remark string `json:"remark" form:"remark" gorm:"column:remark;comment:备注(异常原因等)"`
}

func (InspectResult) TableName() string { return "ops_inspect_results" }
