package ops

import (
	"context"
	"strings"

	"github.com/flipped-aurora/gin-vue-admin/server/global"
	"github.com/flipped-aurora/gin-vue-admin/server/model/ops"
	opsReq "github.com/flipped-aurora/gin-vue-admin/server/model/ops/request"
)

// AuditService 运维操作审计: 记录与查询。
type AuditService struct{}

// maxDetailLen 审计详情最大长度, 防止命令输出过长撑爆表
const maxDetailLen = 2000

// Record 写入一条审计记录(同步落表, 失败仅记日志不阻断主流程由调用方决定)。
func (s *AuditService) Record(ctx context.Context, operatorID uint, operator, action, target, ip, status, detail string) error {
	if len(detail) > maxDetailLen {
		detail = detail[:maxDetailLen]
	}
	r := ops.AuditRecord{
		OperatorID: operatorID,
		Operator:   operator,
		Action:     action,
		Target:     target,
		IP:         ip,
		Status:     status,
		Detail:     detail,
	}
	return global.GVA_DB.WithContext(ctx).Create(&r).Error
}

// GetList 分页查询审计记录
func (s *AuditService) GetList(ctx context.Context, info opsReq.AuditSearch) (list []ops.AuditRecord, total int64, err error) {
	db := global.GVA_DB.WithContext(ctx).Model(&ops.AuditRecord{})
	if info.Action != "" {
		db = db.Where("action = ?", info.Action)
	}
	if info.Status != "" {
		db = db.Where("status = ?", info.Status)
	}
	if info.OperatorID > 0 {
		db = db.Where("operator_id = ?", info.OperatorID)
	}
	if info.Keyword != "" {
		like := "%" + info.Keyword + "%"
		db = db.Where("operator LIKE ? OR target LIKE ? OR detail LIKE ?", like, like, like)
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

// trim 截断辅助
func trim(s string, n int) string {
	s = strings.TrimSpace(s)
	if len(s) > n {
		return s[:n]
	}
	return s
}
