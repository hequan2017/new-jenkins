package ops

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/flipped-aurora/gin-vue-admin/server/global"
	"github.com/flipped-aurora/gin-vue-admin/server/model/ops"
	opsReq "github.com/flipped-aurora/gin-vue-admin/server/model/ops/request"
	"github.com/flipped-aurora/gin-vue-admin/server/utils/datascope"
	"github.com/flipped-aurora/gin-vue-admin/server/utils/logger"
	"github.com/robfig/cron/v3"
	"gorm.io/gorm"
)

// InspectService 巡检任务管理 + 调度执行。
// 调度复用 global.GVA_TIMER, 命令执行复用 SSHService。
type InspectService struct{}

// ValidateSpec 校验 cron 表达式(参考 workflow.ValidateSpec)
func (s *InspectService) ValidateSpec(spec string, withSeconds bool) error {
	var err error
	if withSeconds {
		_, err = cron.NewParser(cron.Second|cron.Minute|cron.Hour|cron.Dom|cron.Month|cron.Dow|cron.Descriptor).Parse(spec)
	} else {
		_, err = cron.ParseStandard(spec)
	}
	if err != nil {
		return fmt.Errorf("cron 表达式非法: %w", err)
	}
	return nil
}

func inspectCronName(id uint) string { return fmt.Sprintf("ops/inspect/%d", id) }

// ScheduleTask 幂等调度: 先 Clear 再按 enabled+spec 重新注册
func (s *InspectService) ScheduleTask(t ops.InspectTask) {
	name := inspectCronName(t.ID)
	global.GVA_Timer.Clear(name)
	if !t.Enabled || t.Spec == "" {
		return
	}
	task := t // 闭包捕获副本
	fn := func() { s.runOnce(task) }
	var err error
	if t.WithSeconds {
		_, err = global.GVA_Timer.AddTaskByFuncWithSecond(name, t.Spec, fn, t.Name)
	} else {
		_, err = global.GVA_Timer.AddTaskByFunc(name, t.Spec, fn, t.Name)
	}
	if err != nil {
		logger.Bg().Mod("ops").Err(err).Error("注册巡检调度失败: " + t.Name)
	}
}

// runOnce 执行一次巡检: SSH 跑命令, 判定关键字, 落结果 + 审计
func (s *InspectService) runOnce(t ops.InspectTask) {
	// 系统上下文(定时任务, 旁路数据权限)
	ctx := datascope.WithSystem(context.Background())
	out, err := (&SSHService{}).ExecCommand(ctx, t.AssetID, t.CredentialID, t.Command)

	status := ops.InspectStatusOK
	remark := ""
	if err != nil {
		status = ops.InspectStatusAlert
		remark = "命令执行失败: " + trim(err.Error(), 500)
	} else if matched, kw := hitKeyword(out, t.Keyword); matched {
		status = ops.InspectStatusAlert
		remark = "命中异常关键字: " + kw
	}
	result := ops.InspectResult{
		TaskID: t.ID,
		Status: status,
		Output: trim(out, 2000),
		Remark: remark,
	}
	if dbErr := global.GVA_DB.WithContext(ctx).Create(&result).Error; dbErr != nil {
		logger.Bg().Mod("ops").Err(dbErr).Error("写入巡检结果失败")
	}
	// 审计(系统操作人)
	_ = (&AuditService{}).Record(ctx, 0, "system", ops.AuditActionInspect,
		fmt.Sprintf("task=%d asset=%d", t.ID, t.AssetID), "", status, trim(out, 500))
}

// hitKeyword 检查输出是否命中任一关键字(逗号分隔)
func hitKeyword(output, keywords string) (bool, string) {
	if strings.TrimSpace(keywords) == "" {
		return false, ""
	}
	for _, k := range strings.Split(keywords, ",") {
		k = strings.TrimSpace(k)
		if k != "" && strings.Contains(output, k) {
			return true, k
		}
	}
	return false, ""
}

// Create 创建巡检任务并按需注册调度
func (s *InspectService) Create(ctx context.Context, in opsReq.InspectTaskInput) (*ops.InspectTask, error) {
	if in.Name == "" {
		return nil, errors.New("任务名称不能为空")
	}
	if in.AssetID == 0 || in.CredentialID == 0 {
		return nil, errors.New("必须选择资产与凭据")
	}
	if in.Command == "" {
		return nil, errors.New("巡检命令不能为空")
	}
	if err := s.ValidateSpec(in.Spec, in.WithSeconds); err != nil {
		return nil, err
	}
	t := ops.InspectTask{
		Name: in.Name, AssetID: in.AssetID, CredentialID: in.CredentialID,
		Command: in.Command, Keyword: in.Keyword, Spec: in.Spec,
		WithSeconds: in.WithSeconds, Enabled: in.Enabled, Remark: in.Remark,
	}
	if err := global.GVA_DB.WithContext(ctx).Create(&t).Error; err != nil {
		return nil, err
	}
	s.ScheduleTask(t)
	return &t, nil
}

// Update 更新巡检任务
func (s *InspectService) Update(ctx context.Context, id uint, in opsReq.InspectTaskInput) error {
	if id == 0 {
		return errors.New("任务 id 不能为空")
	}
	if err := s.ValidateSpec(in.Spec, in.WithSeconds); err != nil {
		return err
	}
	t := ops.InspectTask{
		Name: in.Name, AssetID: in.AssetID, CredentialID: in.CredentialID,
		Command: in.Command, Keyword: in.Keyword, Spec: in.Spec,
		WithSeconds: in.WithSeconds, Enabled: in.Enabled, Remark: in.Remark,
	}
	t.ID = id
	if err := global.GVA_DB.WithContext(ctx).Save(&t).Error; err != nil {
		return err
	}
	s.ScheduleTask(t)
	return nil
}

// Toggle 启停(联动调度)
func (s *InspectService) Toggle(ctx context.Context, id uint, enabled bool) error {
	var t ops.InspectTask
	if err := global.GVA_DB.WithContext(ctx).First(&t, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return errors.New("任务不存在")
		}
		return err
	}
	t.Enabled = enabled
	if err := global.GVA_DB.WithContext(ctx).Save(&t).Error; err != nil {
		return err
	}
	s.ScheduleTask(t)
	return nil
}

func (s *InspectService) Delete(ctx context.Context, id uint) error {
	global.GVA_Timer.Clear(inspectCronName(id))
	if err := global.GVA_DB.WithContext(ctx).Delete(&ops.InspectTask{}, id).Error; err != nil {
		return err
	}
	// 历史结果保留, 不级联删
	return nil
}

func (s *InspectService) GetByID(ctx context.Context, id uint) (*ops.InspectTask, error) {
	var t ops.InspectTask
	if err := global.GVA_DB.WithContext(ctx).First(&t, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("任务不存在")
		}
		return nil, err
	}
	return &t, nil
}

func (s *InspectService) GetTaskList(ctx context.Context, info opsReq.InspectTaskSearch) (list []ops.InspectTask, total int64, err error) {
	db := global.GVA_DB.WithContext(ctx).Model(&ops.InspectTask{})
	if info.Name != "" {
		db = db.Where("name LIKE ?", "%"+info.Name+"%")
	}
	if info.Enabled != nil {
		db = db.Where("enabled = ?", *info.Enabled)
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

func (s *InspectService) GetResultList(ctx context.Context, info opsReq.InspectResultSearch) (list []ops.InspectResult, total int64, err error) {
	db := global.GVA_DB.WithContext(ctx).Model(&ops.InspectResult{})
	if info.TaskID > 0 {
		db = db.Where("task_id = ?", info.TaskID)
	}
	if info.Status != "" {
		db = db.Where("status = ?", info.Status)
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

// RunNow 手动触发一次巡检
func (s *InspectService) RunNow(ctx context.Context, id uint) error {
	t, err := s.GetByID(ctx, id)
	if err != nil {
		return err
	}
	go s.runOnce(*t)
	return nil
}

// LoadSchedules 启动时从 DB 恢复所有启用任务的调度(幂等)
func (s *InspectService) LoadSchedules() {
	ctx := datascope.WithSystem(context.Background())
	var tasks []ops.InspectTask
	if err := global.GVA_DB.WithContext(ctx).Where("enabled = ?", true).Find(&tasks).Error; err != nil {
		logger.Bg().Mod("ops").Err(err).Error("加载巡检调度失败")
		return
	}
	for _, t := range tasks {
		s.ScheduleTask(t)
	}
	logger.Bg().Mod("ops").Field("count", len(tasks)).Info("已恢复巡检调度")
}
