package workflow

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strconv"
	"time"

	"github.com/flipped-aurora/gin-vue-admin/server/global"
	"github.com/flipped-aurora/gin-vue-admin/server/model/workflow"
	workflowReq "github.com/flipped-aurora/gin-vue-admin/server/model/workflow/request"
	workflowRes "github.com/flipped-aurora/gin-vue-admin/server/model/workflow/response"
	"github.com/flipped-aurora/gin-vue-admin/server/utils/datascope"
	"github.com/flipped-aurora/gin-vue-admin/server/utils/logger"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

// BuildService 构建实例: 触发 / 查询 / 取消 / 审批 gate
type BuildService struct{}

// TriggerBuild 创建一条构建记录并异步启动引擎执行
// triggerBy 为触发人用户ID(0 表示系统触发)。返回新建构建ID。
func (s *BuildService) TriggerBuild(ctx context.Context, pipelineID uint, params []workflow.ParamValue, trigger string, triggerBy uint) (buildID uint, err error) {
	var p workflow.Pipeline
	now := time.Now()
	build := workflow.PipelineBuild{
		PipelineID: pipelineID,
		Status:     workflow.BuildStatusRunning,
		Trigger:    trigger,
		TriggerBy:  triggerBy,
		StartedAt:  &now,
	}
	err = global.GVA_DB.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		// 定义读取、参数校验、构建号分配与快照创建共享同一行锁和事务。
		if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).
			Preload("Stages", func(db *gorm.DB) *gorm.DB {
				return db.Order("wf_pipeline_stages.sort_order ASC, id ASC")
			}).
			Preload("Stages.Steps", func(db *gorm.DB) *gorm.DB {
				return db.Order("wf_pipeline_steps.sort_order ASC, id ASC")
			}).
			First(&p, pipelineID).Error; err != nil {
			return fmt.Errorf("流水线不存在: %w", err)
		}
		if !p.Enabled {
			return errors.New("流水线未启用")
		}
		if len(p.Stages) == 0 {
			return errors.New("流水线未配置任何阶段")
		}
		validParams, err := validateAndFillParams(p.ParamSchema, params)
		if err != nil {
			return err
		}
		build.Params = mustJSON(validParams)
		var maxNo int
		if err := tx.Model(&workflow.PipelineBuild{}).
			Where("pipeline_id = ?", pipelineID).
			Select("COALESCE(MAX(build_no), 0)").Scan(&maxNo).Error; err != nil {
			return err
		}
		build.BuildNo = maxNo + 1
		if err := tx.Create(&build).Error; err != nil {
			return err
		}
		// 完整快照阶段行为与步骤配置，定义后续被修改/删除不影响已触发构建。
		for _, st := range p.Stages {
			bs := workflow.PipelineBuildStage{
				BuildID:                 build.ID,
				StageID:                 st.ID,
				SnapshotName:            st.Name,
				SnapshotOrder:           st.Order,
				SnapshotApproval:        st.Approval,
				SnapshotContinueOnError: st.ContinueOnError,
				SnapshotParallel:        st.Parallel,
				Status:                  workflow.BuildStatusPending,
			}
			if err := tx.Create(&bs).Error; err != nil {
				return err
			}
			for _, sp := range st.Steps {
				bstep := workflow.PipelineBuildStep{
					BuildID:        build.ID,
					StageID:        bs.ID,
					StepID:         sp.ID,
					SnapshotName:   sp.Name,
					SnapshotType:   sp.Type,
					SnapshotConfig: append([]byte(nil), sp.Config...),
					SnapshotOrder:  sp.Order,
					Status:         workflow.BuildStatusPending,
				}
				if err := tx.Create(&bstep).Error; err != nil {
					return err
				}
			}
		}
		return nil
	})
	if err != nil {
		return 0, err
	}

	// 4. 异步启动引擎; goroutine 内用系统上下文(datascope.WithSystem), 不裸 context.Background()
	engineCtx := datascope.WithSystem(context.Background())
	go EngineApp.Run(engineCtx, build.ID, triggerBy)
	return build.ID, nil
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
// 更新状态后通知 Engine 取消运行上下文，可立即中断 Shell/HTTP 或审批等待。
func (s *BuildService) CancelBuild(ctx context.Context, id uint) error {
	rows, err := s.setBuildStatusIfIn(ctx, id, workflow.BuildStatusCanceled,
		workflow.BuildStatusRunning, workflow.BuildStatusApproval, workflow.BuildStatusPending)
	if err != nil {
		return err
	}
	if rows == 0 {
		return errors.New("当前构建状态不可取消")
	}
	EngineApp.Cancel(id)
	return nil
}

// RetryBuild 重跑历史构建: 取原 build 的 pipelineID + params, 触发一次新构建(manual 触发)
func (s *BuildService) RetryBuild(ctx context.Context, buildID uint, triggerBy uint) (uint, error) {
	var build workflow.PipelineBuild
	if err := global.GVA_DB.WithContext(ctx).First(&build, buildID).Error; err != nil {
		return 0, fmt.Errorf("构建不存在: %w", err)
	}
	// 解析历史 params
	var params []workflow.ParamValue
	if len(build.Params) > 0 {
		_ = json.Unmarshal(build.Params, &params)
	}
	return s.TriggerBuild(ctx, build.PipelineID, params, workflow.TriggerManual, triggerBy)
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

// validateAndFillParams 对照 ParamSchema 校验入参: 必填检查、类型宽松校验、缺省用 Default 回填。
// 返回规整后的 params(含回填的默认值)。ParamSchema 为空时直接放行(无参流水线)。
func validateAndFillParams(schemaJSON []byte, params []workflow.ParamValue) ([]workflow.ParamValue, error) {
	if len(schemaJSON) == 0 {
		return params, nil
	}
	var schema []workflow.ParamField
	if err := json.Unmarshal(schemaJSON, &schema); err != nil {
		return nil, errors.New("流水线参数定义(paramSchema)格式错误")
	}
	// 入参转 map 便于查找
	inputMap := make(map[string]string, len(params))
	for _, p := range params {
		if _, exists := inputMap[p.Name]; exists {
			return nil, fmt.Errorf("参数重复: %s", p.Name)
		}
		inputMap[p.Name] = p.Value
	}
	// 对照 schema 逐项校验 + 回填默认值
	result := make([]workflow.ParamValue, 0, len(schema))
	schemaNames := make(map[string]struct{}, len(schema))
	for _, f := range schema {
		if f.Name == "" {
			return nil, errors.New("流水线参数定义包含空名称")
		}
		if _, exists := schemaNames[f.Name]; exists {
			return nil, fmt.Errorf("流水线参数定义名称重复: %s", f.Name)
		}
		schemaNames[f.Name] = struct{}{}
		v, ok := inputMap[f.Name]
		if !ok || v == "" {
			if f.Required && f.Default == "" {
				return nil, fmt.Errorf("缺少必填参数: %s", f.Name)
			}
			v = f.Default // 缺省用默认值(可能为空)
		}
		// 类型宽松校验: number 要求能解析为数字, bool 要求 true/false
		if v != "" {
			if err := checkParamType(f.Type, v); err != nil {
				return nil, fmt.Errorf("参数 %s 类型不合法: %w", f.Name, err)
			}
		}
		result = append(result, workflow.ParamValue{Name: f.Name, Value: v})
	}
	for name := range inputMap {
		if _, exists := schemaNames[name]; !exists {
			return nil, fmt.Errorf("未声明的参数: %s", name)
		}
	}
	return result, nil
}

// checkParamType 按 ParamField.Type 做宽松类型校验
func checkParamType(typ, v string) error {
	switch typ {
	case "", "string":
		return nil
	case "number":
		if _, err := strconv.ParseFloat(v, 64); err != nil {
			return fmt.Errorf("期望数字, 实际 %q", v)
		}
	case "bool":
		if v != "true" && v != "false" {
			return fmt.Errorf("期望 bool(true/false), 实际 %q", v)
		}
	default:
		return fmt.Errorf("未知参数类型 %s", typ)
	}
	return nil
}

// parseNumber 别名, 保留语义(实际由 checkParamType 内联 strconv 实现)
func parseNumber(v string) (float64, error) { return strconv.ParseFloat(v, 64) }

// decodeBuildParams 把 build.Params([]ParamValue JSON) 解析为 map[name]value,
// 供执行器做 ${param.xxx} 变量替换。解析失败返回空 map(降级,不阻断执行)。
func decodeBuildParams(raw []byte) map[string]string {
	out := map[string]string{}
	if len(raw) == 0 {
		return out
	}
	var pvs []workflow.ParamValue
	if err := json.Unmarshal(raw, &pvs); err != nil {
		return out
	}
	for _, p := range pvs {
		out[p.Name] = p.Value
	}
	return out
}
