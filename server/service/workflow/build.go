package workflow

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/flipped-aurora/gin-vue-admin/server/global"
	"github.com/flipped-aurora/gin-vue-admin/server/model/workflow"
	workflowReq "github.com/flipped-aurora/gin-vue-admin/server/model/workflow/request"
	workflowRes "github.com/flipped-aurora/gin-vue-admin/server/model/workflow/response"
	"github.com/flipped-aurora/gin-vue-admin/server/utils/datascope"
	"github.com/flipped-aurora/gin-vue-admin/server/utils/logger"
	"gorm.io/gorm"
)

// BuildService 构建实例: 触发 / 查询 / 取消 / 审批 gate
type BuildService struct{}

// TriggerBuild 创建一条构建记录并异步启动引擎执行
// triggerBy 为触发人用户ID(0 表示系统触发)。返回新建构建ID。
func (s *BuildService) TriggerBuild(ctx context.Context, pipelineID uint, params []workflow.ParamValue, trigger string, triggerBy uint) (buildID uint, err error) {
	// 1. 取流水线定义(含 Stage/Step 树), 校验启用状态
	var p workflow.Pipeline
	if err = global.GVA_DB.WithContext(ctx).
		Preload("Stages", func(db *gorm.DB) *gorm.DB {
			return db.Order("wf_pipeline_stages.sort_order ASC, id ASC")
		}).
		Preload("Stages.Steps", func(db *gorm.DB) *gorm.DB {
			return db.Order("wf_pipeline_steps.sort_order ASC, id ASC")
		}).
		First(&p, pipelineID).Error; err != nil {
		return 0, fmt.Errorf("流水线不存在: %w", err)
	}
	if !p.Enabled {
		return 0, errors.New("流水线未启用")
	}
	if len(p.Stages) == 0 {
		return 0, errors.New("流水线未配置任何阶段")
	}

	// 2. 计算 buildNo(同流水线下自增)
	buildNo, err := s.nextBuildNo(ctx, pipelineID)
	if err != nil {
		return 0, err
	}

	// 3. 创建 build + build_stage + build_step 运行实例
	now := time.Now()
	build := workflow.PipelineBuild{
		PipelineID: pipelineID,
		BuildNo:    buildNo,
		Status:     workflow.BuildStatusRunning,
		Params:     mustJSON(params),
		Trigger:    trigger,
		TriggerBy:  triggerBy,
		StartedAt:  &now,
	}
	if err = global.GVA_DB.WithContext(ctx).Create(&build).Error; err != nil {
		return 0, err
	}
	// 快照阶段 / 步骤定义到运行实例(定义后续被改不影响历史)
	for _, st := range p.Stages {
		bs := workflow.PipelineBuildStage{
			BuildID:       build.ID,
			StageID:       st.ID,
			SnapshotName:  st.Name,
			SnapshotOrder: st.Order,
			Status:        workflow.BuildStatusPending,
		}
		if err = global.GVA_DB.WithContext(ctx).Create(&bs).Error; err != nil {
			return 0, err
		}
		for _, sp := range st.Steps {
			bstep := workflow.PipelineBuildStep{
				BuildID:       build.ID,
				StageID:       bs.ID,
				StepID:        sp.ID,
				SnapshotName:  sp.Name,
				SnapshotType:  sp.Type,
				SnapshotOrder: sp.Order,
				Status:        workflow.BuildStatusPending,
			}
			if err = global.GVA_DB.WithContext(ctx).Create(&bstep).Error; err != nil {
				return 0, err
			}
		}
	}

	// 4. 异步启动引擎; goroutine 内用系统上下文(datascope.WithSystem), 不裸 context.Background()
	engineCtx := datascope.WithSystem(context.Background())
	go EngineApp.Run(engineCtx, build.ID, triggerBy)
	return build.ID, nil
}

// nextBuildNo 同流水线下构建序号自增(并发安全: 用子查询取 MAX+1)
func (s *BuildService) nextBuildNo(ctx context.Context, pipelineID uint) (int, error) {
	var maxNo int
	err := global.GVA_DB.WithContext(ctx).
		Model(&workflow.PipelineBuild{}).
		Where("pipeline_id = ?", pipelineID).
		Select("COALESCE(MAX(build_no), 0)").Scan(&maxNo).Error
	if err != nil {
		return 0, err
	}
	return maxNo + 1, nil
}

// GetBuildList 分页查询构建历史
func (s *BuildService) GetBuildList(ctx context.Context, info workflowReq.PipelineBuildSearch) (list []workflow.PipelineBuild, total int64, err error) {
	limit, offset := info.LimitOffset()
	if limit == 0 {
		limit = 10
	}
	db := global.GVA_DB.WithContext(ctx).Model(&workflow.PipelineBuild{})
	if info.PipelineID > 0 {
		db = db.Where("pipeline_id = ?", info.PipelineID)
	}
	if info.Status != "" {
		db = db.Where("status = ?", info.Status)
	}
	if info.Trigger != "" {
		db = db.Where("trigger = ?", info.Trigger)
	}
	if err = db.Count(&total).Error; err != nil {
		return
	}
	err = db.Limit(limit).Offset(offset).Order("id desc").Find(&list).Error
	return
}

// GetBuildDetail 构建详情(含 stage/step 运行视图)
func (s *BuildService) GetBuildDetail(ctx context.Context, id uint) (detail workflowRes.BuildDetail, err error) {
	if err = global.GVA_DB.WithContext(ctx).First(&detail.Build, id).Error; err != nil {
		return
	}
	if err = global.GVA_DB.WithContext(ctx).
		Where("build_id = ?", id).
		Order("snapshot_order ASC, id ASC").
		Find(&detail.Stages).Error; err != nil {
		return
	}
	if err = global.GVA_DB.WithContext(ctx).
		Where("build_id = ?", id).
		Order("snapshot_order ASC, id ASC").
		Find(&detail.Steps).Error; err != nil {
		return
	}
	return
}

// GetBuildLog 分页拉历史日志(按 step 维度)
func (s *BuildService) GetBuildLog(ctx context.Context, info workflowReq.BuildLogSearch) (list []workflow.PipelineBuildLog, total int64, err error) {
	limit, offset := info.LimitOffset()
	if limit == 0 {
		limit = 200 // 日志默认拉多一些
	}
	db := global.GVA_DB.WithContext(ctx).Model(&workflow.PipelineBuildLog{})
	if info.BuildID > 0 {
		db = db.Where("build_id = ?", info.BuildID)
	}
	if info.StepID > 0 {
		db = db.Where("step_id = ?", info.StepID)
	}
	if err = db.Count(&total).Error; err != nil {
		return
	}
	err = db.Limit(limit).Offset(offset).Order("id ASC").Find(&list).Error
	return
}

// CancelBuild 取消构建: 仅 running / running-approval / pending 可取消
// 实际中断由 Engine 检查 ctx 实现; 这里把状态标记为 canceled, Engine 轮询到即停。
func (s *BuildService) CancelBuild(ctx context.Context, id uint) error {
	rows, err := s.setBuildStatusIfIn(ctx, id, workflow.BuildStatusCanceled,
		workflow.BuildStatusRunning, workflow.BuildStatusApproval, workflow.BuildStatusPending)
	if err != nil {
		return err
	}
	if rows == 0 {
		return errors.New("当前构建状态不可取消")
	}
	return nil
}

// ApproveStage 审批 gate: approve=true 继续下一阶段, false 标记构建失败
func (s *BuildService) ApproveStage(ctx context.Context, req workflowReq.ApproveStageReq) error {
	var build workflow.PipelineBuild
	if err := global.GVA_DB.WithContext(ctx).First(&build, req.BuildID).Error; err != nil {
		return err
	}
	if build.Status != workflow.BuildStatusApproval {
		return errors.New("当前构建不在等待审批状态")
	}
	// 用 approval 通道唤醒 Engine(见 engine.go 的 ApproveChan)
	if req.Approve {
		EngineApp.NotifyApproval(req.BuildID, true, req.Comment)
		return nil
	}
	// 拒绝: 直接标记 failed 并唤醒(让 Engine 退出等待)
	EngineApp.NotifyApproval(req.BuildID, false, req.Comment)
	return nil
}

// setBuildStatusIfIn 仅当 build 当前状态属于 inSet 之一时才更新为 newStatus, 返回受影响行数
func (s *BuildService) setBuildStatusIfIn(ctx context.Context, id uint, newStatus string, inSet ...string) (int64, error) {
	if len(inSet) == 0 {
		return 0, nil
	}
	now := time.Now()
	db := global.GVA_DB.WithContext(ctx).Model(&workflow.PipelineBuild{}).
		Where("id = ?", id).
		Where("status IN ?", inSet).
		Updates(map[string]interface{}{
			"status":      newStatus,
			"finished_at": now,
		})
	return db.RowsAffected, db.Error
}

// logErr 引擎内部统一记错(避免在每个分支重复写 logger 调用)
func logErr(buildID uint, msg string, err error) {
	logger.Bg().Mod("workflow").Err(err).Error(fmt.Sprintf("build %d: %s", buildID, msg))
}
