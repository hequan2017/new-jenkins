package ops

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/flipped-aurora/gin-vue-admin/server/global"
	"github.com/flipped-aurora/gin-vue-admin/server/model/ops"
	opsReq "github.com/flipped-aurora/gin-vue-admin/server/model/ops/request"
	"github.com/flipped-aurora/gin-vue-admin/server/model/workflow"
	"gorm.io/datatypes"
	"gorm.io/gorm"
)

// TicketService 工单发版: 申请 -> 审批 -> 触发流水线构建。
type TicketService struct{}

// Create 创建工单(状态 pending)。params 为触发流水线的入参列表。
func (s *TicketService) Create(ctx context.Context, in opsReq.TicketInput, applicantID uint) (*ops.Ticket, error) {
	if in.Title == "" {
		return nil, errors.New("工单标题不能为空")
	}
	if in.PipelineID == 0 {
		return nil, errors.New("必须选择一条流水线")
	}
	raw, _ := json.Marshal(in.Params)
	t := ops.Ticket{
		Title:       in.Title,
		PipelineID:  in.PipelineID,
		Params:      datatypes.JSON(raw),
		Status:      ops.TicketStatusPending,
		ApplicantID: applicantID,
		ApplyReason: in.ApplyReason,
		Remark:      in.Remark,
	}
	if err := global.GVA_DB.WithContext(ctx).Create(&t).Error; err != nil {
		return nil, err
	}
	return &t, nil
}

func (s *TicketService) Update(ctx context.Context, t *ops.Ticket) error {
	return global.GVA_DB.WithContext(ctx).Save(t).Error
}

func (s *TicketService) Delete(ctx context.Context, id uint) error {
	return global.GVA_DB.WithContext(ctx).Delete(&ops.Ticket{}, id).Error
}

func (s *TicketService) GetByID(ctx context.Context, id uint) (*ops.Ticket, error) {
	var t ops.Ticket
	if err := global.GVA_DB.WithContext(ctx).First(&t, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("工单不存在")
		}
		return nil, err
	}
	return &t, nil
}

// GetList 分页查询工单。applicantID > 0 时仅返回该申请人的工单。
func (s *TicketService) GetList(ctx context.Context, info opsReq.TicketSearch) (list []ops.Ticket, total int64, err error) {
	db := global.GVA_DB.WithContext(ctx).Model(&ops.Ticket{})
	if info.Status != "" {
		db = db.Where("status = ?", info.Status)
	}
	if info.ApplicantID > 0 {
		db = db.Where("applicant_id = ?", info.ApplicantID)
	}
	if info.PipelineID > 0 {
		db = db.Where("pipeline_id = ?", info.PipelineID)
	}
	if info.Keyword != "" {
		db = db.Where("title LIKE ?", "%"+info.Keyword+"%")
	}
	if err = db.Count(&total).Error; err != nil {
		return
	}
	limit, offset := info.LimitOffset()
	if limit == 0 {
		limit = 10
	}
	err = db.Order("id DESC").Limit(limit).Offset(offset).Find(&list).Error
	return
}

// Approve 审批工单。approve=true 通过并触发流水线; false 拒绝。
func (s *TicketService) Approve(ctx context.Context, req opsReq.ApproveTicketReq, approverID uint) (*ops.Ticket, error) {
	t, err := s.GetByID(ctx, req.ID)
	if err != nil {
		return nil, err
	}
	if t.Status != ops.TicketStatusPending {
		return nil, fmt.Errorf("当前状态(%s)不可审批", t.Status)
	}
	t.ApproverID = approverID
	t.ApproveComment = req.Comment

	if !req.Approve {
		t.Status = ops.TicketStatusRejected
		if err := global.GVA_DB.WithContext(ctx).Save(t).Error; err != nil {
			return nil, err
		}
		return t, nil
	}

	// 通过: 触发绑定的流水线构建, 复用 workflow.BuildService.TriggerBuild(经注入, 避免循环导入)
	params, err := decodeParams(t.Params)
	if err != nil {
		return nil, fmt.Errorf("解析工单参数失败: %w", err)
	}
	if WorkflowTriggerProvider == nil {
		return nil, errors.New("流水线触发能力未初始化")
	}
	buildID, err := WorkflowTriggerProvider.TriggerBuild(
		ctx, t.PipelineID, params, ops.TriggerTicket, approverID,
	)
	if err != nil {
		// 触发失败: 标记 failed 并保留原因, 不抛出原始错误细节给前端过多信息
		t.Status = ops.TicketStatusFailed
		t.ApproveComment = appendComment(req.Comment, "触发失败: "+err.Error())
		_ = global.GVA_DB.WithContext(ctx).Save(t).Error
		return t, fmt.Errorf("触发流水线失败: %w", err)
	}

	t.Status = ops.TicketStatusDeploying
	t.BuildID = buildID
	if err := global.GVA_DB.WithContext(ctx).Save(t).Error; err != nil {
		return t, err
	}
	return t, nil
}

// SyncBuildStatus 根据构建最终状态回写工单终态(成功 / 失败 / 取消)。
// 由调用方在拉取构建详情后调用, 或留给后续轮询 / 事件订阅。
func (s *TicketService) SyncBuildStatus(ctx context.Context, ticketID uint, buildStatus string) error {
	t, err := s.GetByID(ctx, ticketID)
	if err != nil {
		return err
	}
	if t.Status != ops.TicketStatusDeploying {
		return nil
	}
	switch buildStatus {
	case workflow.BuildStatusSuccess:
		t.Status = ops.TicketStatusSuccess
	case workflow.BuildStatusFailed:
		t.Status = ops.TicketStatusFailed
	case workflow.BuildStatusCanceled:
		t.Status = ops.TicketStatusCanceled
	default:
		return nil
	}
	return global.GVA_DB.WithContext(ctx).Save(t).Error
}

// Cancel 取消工单(仅 pending 可取消)。
func (s *TicketService) Cancel(ctx context.Context, id uint) error {
	t, err := s.GetByID(ctx, id)
	if err != nil {
		return err
	}
	if t.Status != ops.TicketStatusPending {
		return fmt.Errorf("当前状态(%s)不可取消", t.Status)
	}
	t.Status = ops.TicketStatusCanceled
	return global.GVA_DB.WithContext(ctx).Save(t).Error
}

// decodeParams 把前端 [{name,value}] 结构解析为 workflow.ParamValue 列表
func decodeParams(raw datatypes.JSON) ([]workflow.ParamValue, error) {
	if len(raw) == 0 {
		return nil, nil
	}
	// 前端传 map 数组, 这里宽松兼容 name/value 与 Name/Value 大小写
	var generic []map[string]interface{}
	if err := json.Unmarshal(raw, &generic); err != nil {
		return nil, err
	}
	out := make([]workflow.ParamValue, 0, len(generic))
	for _, m := range generic {
		name, _ := m["name"].(string)
		if name == "" {
			name, _ = m["Name"].(string)
		}
		var val string
		switch v := m["value"].(type) {
		case string:
			val = v
		default:
			if v != nil {
				b, _ := json.Marshal(v)
				val = string(b)
			}
		}
		out = append(out, workflow.ParamValue{Name: name, Value: val})
	}
	return out, nil
}

func appendComment(old, add string) string {
	if old == "" {
		return add
	}
	return old + "\n" + add
}
