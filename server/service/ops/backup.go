package ops

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/flipped-aurora/gin-vue-admin/server/global"
	"github.com/flipped-aurora/gin-vue-admin/server/model/ops"
	opsReq "github.com/flipped-aurora/gin-vue-admin/server/model/ops/request"
	"github.com/flipped-aurora/gin-vue-admin/server/utils/datascope"
	"github.com/flipped-aurora/gin-vue-admin/server/utils/logger"
	"github.com/robfig/cron/v3"
	"gorm.io/gorm"
)

// BackupService 备份任务: 定期 SSH+SFTP 拉取远程文件归档到本地 data/backups/。
type BackupService struct{}

const backupRoot = "data/backups"

func (s *BackupService) ValidateSpec(spec string, withSeconds bool) error {
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

func backupCronName(id uint) string { return fmt.Sprintf("ops/backup/%d", id) }

func (s *BackupService) ScheduleTask(t ops.BackupTask) {
	name := backupCronName(t.ID)
	global.GVA_Timer.Clear(name)
	if !t.Enabled || t.Spec == "" {
		return
	}
	task := t
	fn := func() { s.runOnce(task) }
	var err error
	if t.WithSeconds {
		_, err = global.GVA_Timer.AddTaskByFuncWithSecond(name, t.Spec, fn, t.Name)
	} else {
		_, err = global.GVA_Timer.AddTaskByFunc(name, t.Spec, fn, t.Name)
	}
	if err != nil {
		logger.Bg().Mod("ops").Err(err).Error("注册备份调度失败: " + t.Name)
	}
}

// runOnce 执行一次备份: 通过 SFTP 拉取远程文件到本地归档目录
func (s *BackupService) runOnce(t ops.BackupTask) {
	ctx := datascope.WithSystem(context.Background())
	rec := ops.BackupRecord{TaskID: t.ID, Status: ops.BackupStatusRunning}
	global.GVA_DB.WithContext(ctx).Create(&rec)

	localDir := filepath.Join(backupRoot, fmt.Sprintf("task-%d", t.ID))
	if mkErr := os.MkdirAll(localDir, 0o755); mkErr != nil {
		s.finishFail(ctx, &rec, "创建本地目录失败: "+mkErr.Error())
		return
	}
	stamp := time.Now().Format("20060102-150405")
	base := filepath.Base(t.RemotePath)
	if base == "" || base == "." || base == "/" {
		base = "root"
	}
	localPath := filepath.Join(localDir, fmt.Sprintf("%s_%s", stamp, base))

	size, err := s.pullRemote(ctx, t, localPath)
	if err != nil {
		s.finishFail(ctx, &rec, "拉取失败: "+err.Error())
		_ = (&AlertService{}).Raise(ctx, ops.AlertSourceBackup, ops.AlertLevelCritical,
			"备份失败: "+t.Name, err.Error(), t.ID, t.Name)
		return
	}
	rec.Status = ops.BackupStatusSuccess
	rec.LocalPath = localPath
	rec.Size = size
	global.GVA_DB.WithContext(ctx).Save(&rec)
	s.pruneOld(ctx, t)
	_ = (&AlertService{}).Resolve(ctx, ops.AlertSourceBackup, t.ID, "备份成功")
}

func (s *BackupService) finishFail(ctx context.Context, rec *ops.BackupRecord, msg string) {
	rec.Status = ops.BackupStatusFailed
	rec.Detail = msg
	global.GVA_DB.WithContext(ctx).Save(rec)
}

// pullRemote 用 SFTP 把远程文件复制到本地
func (s *BackupService) pullRemote(ctx context.Context, t ops.BackupTask, localPath string) (int64, error) {
	f, err := os.Create(localPath)
	if err != nil {
		return 0, err
	}
	defer f.Close()

	// 复用 FileService 的 SFTP 能力: 直接读取远程文件内容
	content, err := (&FileService{}).ReadFile(ctx, t.AssetID, t.CredentialID, t.RemotePath)
	if err != nil {
		return 0, err
	}
	n, err := f.WriteString(content)
	if err != nil {
		return 0, err
	}
	return int64(n), nil
}

// pruneOld 按保留份数清理最旧记录与本地文件
func (s *BackupService) pruneOld(ctx context.Context, t ops.BackupTask) {
	keep := t.KeepCount
	if keep <= 0 {
		keep = 7
	}
	var recs []ops.BackupRecord
	global.GVA_DB.WithContext(ctx).
		Where("task_id = ? AND status = ?", t.ID, ops.BackupStatusSuccess).
		Order("id DESC").Find(&recs)
	if len(recs) <= keep {
		return
	}
	for _, r := range recs[keep:] {
		if r.LocalPath != "" {
			_ = os.Remove(r.LocalPath)
		}
		global.GVA_DB.WithContext(ctx).Delete(&r)
	}
}

func (s *BackupService) Create(ctx context.Context, in opsReq.BackupTaskInput) (*ops.BackupTask, error) {
	if in.Name == "" || in.RemotePath == "" {
		return nil, errors.New("名称与远程路径不能为空")
	}
	if err := s.ValidateSpec(in.Spec, in.WithSeconds); err != nil {
		return nil, err
	}
	t := ops.BackupTask{
		Name: in.Name, AssetID: in.AssetID, CredentialID: in.CredentialID,
		RemotePath: in.RemotePath, Spec: in.Spec, WithSeconds: in.WithSeconds,
		Enabled: in.Enabled, KeepCount: in.KeepCount, Remark: in.Remark,
	}
	if t.KeepCount == 0 {
		t.KeepCount = 7
	}
	if err := global.GVA_DB.WithContext(ctx).Create(&t).Error; err != nil {
		return nil, err
	}
	s.ScheduleTask(t)
	return &t, nil
}

func (s *BackupService) Update(ctx context.Context, id uint, in opsReq.BackupTaskInput) error {
	if id == 0 {
		return errors.New("任务 id 不能为空")
	}
	if err := s.ValidateSpec(in.Spec, in.WithSeconds); err != nil {
		return err
	}
	t := ops.BackupTask{
		Name: in.Name, AssetID: in.AssetID, CredentialID: in.CredentialID,
		RemotePath: in.RemotePath, Spec: in.Spec, WithSeconds: in.WithSeconds,
		Enabled: in.Enabled, KeepCount: in.KeepCount, Remark: in.Remark,
	}
	t.ID = id
	if t.KeepCount == 0 {
		t.KeepCount = 7
	}
	if err := global.GVA_DB.WithContext(ctx).Save(&t).Error; err != nil {
		return err
	}
	s.ScheduleTask(t)
	return nil
}

func (s *BackupService) Toggle(ctx context.Context, id uint, enabled bool) error {
	var t ops.BackupTask
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

func (s *BackupService) Delete(ctx context.Context, id uint) error {
	global.GVA_Timer.Clear(backupCronName(id))
	return global.GVA_DB.WithContext(ctx).Delete(&ops.BackupTask{}, id).Error
}

func (s *BackupService) GetByID(ctx context.Context, id uint) (*ops.BackupTask, error) {
	var t ops.BackupTask
	if err := global.GVA_DB.WithContext(ctx).First(&t, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("任务不存在")
		}
		return nil, err
	}
	return &t, nil
}

func (s *BackupService) GetTaskList(ctx context.Context, info opsReq.BackupTaskSearch) (list []ops.BackupTask, total int64, err error) {
	db := global.GVA_DB.WithContext(ctx).Model(&ops.BackupTask{})
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

func (s *BackupService) GetRecordList(ctx context.Context, info opsReq.BackupRecordSearch) (list []ops.BackupRecord, total int64, err error) {
	db := global.GVA_DB.WithContext(ctx).Model(&ops.BackupRecord{})
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

// RunNow 手动触发一次备份
func (s *BackupService) RunNow(ctx context.Context, id uint) error {
	t, err := s.GetByID(ctx, id)
	if err != nil {
		return err
	}
	go s.runOnce(*t)
	return nil
}

// ReadArchive 读取本地归档文件内容(供下载/恢复), 返回文件路径
func (s *BackupService) ReadArchive(recordID uint) (string, error) {
	var rec ops.BackupRecord
	if err := global.GVA_DB.First(&rec, recordID).Error; err != nil {
		return "", err
	}
	if rec.LocalPath == "" {
		return "", errors.New("该记录无本地归档文件")
	}
	if _, err := os.Stat(rec.LocalPath); err != nil {
		return "", errors.New("归档文件不存在: " + rec.LocalPath)
	}
	return rec.LocalPath, nil
}

// LoadSchedules 启动时恢复启用的备份任务调度
func (s *BackupService) LoadSchedules() {
	ctx := datascope.WithSystem(context.Background())
	var tasks []ops.BackupTask
	if err := global.GVA_DB.WithContext(ctx).Where("enabled = ?", true).Find(&tasks).Error; err != nil {
		logger.Bg().Mod("ops").Err(err).Error("加载备份调度失败")
		return
	}
	for _, t := range tasks {
		s.ScheduleTask(t)
	}
	logger.Bg().Mod("ops").Field("count", len(tasks)).Info("已恢复备份调度")
}
