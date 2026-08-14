package ops

import (
	"github.com/flipped-aurora/gin-vue-admin/server/global"
)

// 审计动作类型
const (
	AuditActionLogin       = "login"        // 登录
	AuditActionCmdExec     = "cmd_exec"     // 跳板机命令执行
	AuditActionTerminal    = "terminal"     // 跳板机终端会话
	AuditActionFileOp      = "file_op"      // SFTP 文件操作
	AuditActionTicketOp    = "ticket_op"    // 工单操作
	AuditActionInspect     = "inspect"      // 巡检
	AuditActionAssetOp     = "asset_op"     // 资产操作
	AuditActionCredentialOp = "credential_op" // 凭据操作
)

// AuditRecord 运维操作审计记录。
// 由跳板机命令执行、工单审批、文件操作、巡检等关键动作落表, 提供可追溯性。
// 不含数据权限列: 审计记录是全局可查询的安全数据。
type AuditRecord struct {
	global.GVA_MODEL
	OperatorID uint   `json:"operatorId" form:"operatorId" gorm:"index;column:operator_id;comment:操作人ID"`
	Operator   string `json:"operator" form:"operator" gorm:"column:operator;comment:操作人用户名"`
	Action     string `json:"action" form:"action" gorm:"index;column:action;comment:动作类型"`
	Target     string `json:"target" form:"target" gorm:"column:target;comment:操作对象(如 主机/工单/文件路径)"`
	IP         string `json:"ip" form:"ip" gorm:"column:ip;comment:来源IP"`
	Status     string `json:"status" form:"status" gorm:"column:status;comment:结果 success|failed"`
	Detail     string `json:"detail" form:"detail" gorm:"column:detail;type:text;comment:详情(命令/操作内容, 截断存储)"`
}

func (AuditRecord) TableName() string { return "ops_audit_records" }
