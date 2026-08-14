package ops

import (
	"context"
	"errors"
	"fmt"

	"github.com/flipped-aurora/gin-vue-admin/server/global"
	"github.com/flipped-aurora/gin-vue-admin/server/model/ops"
	opsReq "github.com/flipped-aurora/gin-vue-admin/server/model/ops/request"
	"gorm.io/gorm"
)

// AlertService 告警中心: 统一接收与处理各类运维告警事件。
type AlertService struct{}

// Raise 产生一条告警(供巡检/工单/备份等调用), 同源同对象且仍 active 的不重复产生。
func (s *AlertService) Raise(ctx context.Context, source, level, title, detail string, refID uint, refName string) error {
	// 去重: 同来源同对象仍有 active 告警则更新详情而非新建
	var existing ops.Alert
	err := global.GVA_DB.WithContext(ctx).
		Where("source = ? AND ref_id = ? AND status = ?", source, refID, ops.AlertStatusActive).
		First(&existing).Error
	if err == nil {
		existing.Detail = detail
		existing.Level = level
		return global.GVA_DB.WithContext(ctx).Save(&existing).Error
	}
	if !errors.Is(err, gorm.ErrRecordNotFound) {
		return err
	}
	a := ops.Alert{
		Source:  source,
		Level:   level,
		Title:   title,
		Detail:  detail,
		RefID:   refID,
		RefName: refName,
		Status:  ops.AlertStatusActive,
	}
	if a.Level == "" {
		a.Level = ops.AlertLevelWarning
	}
	return global.GVA_DB.WithContext(ctx).Create(&a).Error
}

// Resolve 自动处理同源同对象的告警(如巡检恢复、工单结束)
func (s *AlertService) Resolve(ctx context.Context, source string, refID uint, comment string) error {
	return global.GVA_DB.WithContext(ctx).Model(&ops.Alert{}).
		Where("source = ? AND ref_id = ? AND status = ?", source, refID, ops.AlertStatusActive).
		Updates(map[string]interface{}{"status": ops.AlertStatusResolved, "comment": comment}).Error
}

func (s *AlertService) GetList(ctx context.Context, info opsReq.AlertSearch) (list []ops.Alert, total int64, err error) {
	db := global.GVA_DB.WithContext(ctx).Model(&ops.Alert{})
	if info.Source != "" {
		db = db.Where("source = ?", info.Source)
	}
	if info.Level != "" {
		db = db.Where("level = ?", info.Level)
	}
	if info.Status != "" {
		db = db.Where("status = ?", info.Status)
	}
	if info.Keyword != "" {
		like := "%" + info.Keyword + "%"
		db = db.Where("title LIKE ? OR ref_name LIKE ?", like, like)
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

// Handle 手动处理告警(标记已处理/忽略)
func (s *AlertService) Handle(ctx context.Context, req opsReq.HandleAlertReq, handler string) error {
	if req.Status != ops.AlertStatusResolved && req.Status != ops.AlertStatusIgnored {
		return fmt.Errorf("非法状态: %s", req.Status)
	}
	return global.GVA_DB.WithContext(ctx).Model(&ops.Alert{}).Where("id = ?", req.ID).
		Updates(map[string]interface{}{"status": req.Status, "comment": req.Comment, "handler": handler}).Error
}

// CountActive 统计未处理告警数(供大盘)
func (s *AlertService) CountActive(ctx context.Context) int64 {
	var cnt int64
	global.GVA_DB.WithContext(ctx).Model(&ops.Alert{}).Where("status = ?", ops.AlertStatusActive).Count(&cnt)
	return cnt
}
