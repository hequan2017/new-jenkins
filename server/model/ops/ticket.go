package ops

import (
	"github.com/flipped-aurora/gin-vue-admin/server/global"
	"gorm.io/datatypes"
)

// 工单状态
const (
	TicketStatusPending   = "pending"   // 待审批
	TicketStatusApproved  = "approved"  // 已通过(触发部署中或已完成)
	TicketStatusRejected  = "rejected"  // 已拒绝
	TicketStatusDeploying = "deploying" // 部署中
	TicketStatusSuccess   = "success"   // 部署成功
	TicketStatusFailed    = "failed"    // 部署失败
	TicketStatusCanceled  = "canceled"  // 已取消
)

// 触发来源: 工单审批通过后触发流水线构建时使用的 trigger 标记, 仅作审计。
const TriggerTicket = "ticket"

// Ticket 工单发版: 申请人选择一条流水线 + 入参, 审批通过后服务端触发构建。
type Ticket struct {
	global.GVA_MODEL
	Title        string         `json:"title" form:"title" gorm:"index;column:title;comment:工单标题"`
	PipelineID   uint           `json:"pipelineId" form:"pipelineId" gorm:"column:pipeline_id;comment:绑定流水线ID"`
	Params       datatypes.JSON `json:"params" form:"params" gorm:"column:params;comment:触发参数(JSON 数组)" swaggertype:"array,object"`
	Status       string         `json:"status" form:"status" gorm:"index;column:status;comment:状态;default:pending"`
	ApplicantID  uint           `json:"applicantId" form:"applicantId" gorm:"column:applicant_id;comment:申请人ID"`
	ApproverID   uint           `json:"approverId" form:"approverId" gorm:"column:approver_id;comment:审批人ID"`
	ApplyReason  string         `json:"applyReason" form:"applyReason" gorm:"column:apply_reason;comment:申请说明"`
	ApproveComment string       `json:"approveComment" form:"approveComment" gorm:"column:approve_comment;comment:审批意见"`
	BuildID      uint           `json:"buildId" form:"buildId" gorm:"column:build_id;comment:触发后回填的构建ID"`
	Remark       string         `json:"remark" form:"remark" gorm:"column:remark;comment:备注"`
}

func (Ticket) TableName() string { return "ops_tickets" }
