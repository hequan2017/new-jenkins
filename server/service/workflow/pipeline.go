package workflow

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/flipped-aurora/gin-vue-admin/server/global"
	"github.com/flipped-aurora/gin-vue-admin/server/model/workflow"
	workflowReq "github.com/flipped-aurora/gin-vue-admin/server/model/workflow/request"
	"gorm.io/datatypes"
	"gorm.io/gorm"
)

// PipelineService 流水线定义 CRUD + Stage/Step 级联保存
type PipelineService struct{}

// CreatePipeline 创建流水线(含 Stage/Step 树, 级联写入)
// 不带数据权限公共字段: 流水线定义是全局共享资源, 不按部门做行级过滤。
func (s *PipelineService) CreatePipeline(ctx context.Context, p *workflow.Pipeline) error {
	if err := s.validatePipeline(p); err != nil {
		return err
	}
	// 级联创建: GORM 默认对 has-many 关联在 Create 时一并处理
	return global.GVA_DB.WithContext(ctx).Create(p).Error
}

// UpdatePipeline 全量覆盖式更新(含 Stage/Step 树)
// 用事务保证一致性: 替换 Stages/Steps, 再更新 Pipeline 本身。
// 归属列(dept_id/created_by)本表没有, 无需 Omit。
func (s *PipelineService) UpdatePipeline(ctx context.Context, p *workflow.Pipeline) error {
	if p.ID == 0 {
		return errors.New("pipeline id 不能为空")
	}
	if err := s.validatePipeline(p); err != nil {
		return err
	}

	return global.GVA_DB.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		// 先清旧 Stage(级联清 Step, 靠外键 OnDelete:CASCADE), 再写新树
		if err := tx.Where("pipeline_id = ?", p.ID).Delete(&workflow.PipelineStage{}).Error; err != nil {
			return err
		}
		// 全量更新 Pipeline 本身(含 ParamSchema)
		if err := tx.Save(p).Error; err != nil {
			return err
		}
		return nil
	})
}

// DeletePipeline 删除流水线(级联删 Stage/Step 定义; 已有 Build 历史保留)
func (s *PipelineService) DeletePipeline(ctx context.Context, id uint) error {
	return global.GVA_DB.WithContext(ctx).Delete(&workflow.Pipeline{}, id).Error
}

// FindPipeline 查流水线详情(含 Stage/Step 树)
func (s *PipelineService) FindPipeline(ctx context.Context, id uint) (p workflow.Pipeline, err error) {
	err = global.GVA_DB.WithContext(ctx).
		Preload("Stages", func(db *gorm.DB) *gorm.DB {
			return db.Order("wf_pipeline_stages.sort_order ASC, wf_pipeline_stages.id ASC")
		}).
		Preload("Stages.Steps", func(db *gorm.DB) *gorm.DB {
			return db.Order("wf_pipeline_steps.sort_order ASC, wf_pipeline_steps.id ASC")
		}).
		First(&p, id).Error
	return
}

// GetPipelineList 分页查询流水线定义列表
func (s *PipelineService) GetPipelineList(ctx context.Context, info workflowReq.PipelineSearch) (list []workflow.Pipeline, total int64, err error) {
	limit, offset := info.LimitOffset()
	if limit == 0 { // PageSize=0 表示不分页查全量(参考 PageInfo.LimitOffset 语义)
		limit = 10
	}
	db := global.GVA_DB.WithContext(ctx).Model(&workflow.Pipeline{})
	if info.Name != "" {
		db = db.Where("name LIKE ?", "%"+info.Name+"%")
	}
	if info.TriggerType != "" {
		db = db.Where("trigger_type = ?", info.TriggerType)
	}
	if info.Enabled != nil {
		db = db.Where("enabled = ?", *info.Enabled)
	}
	if err = db.Count(&total).Error; err != nil {
		return
	}
	err = db.Limit(limit).Offset(offset).Order("id desc").Find(&list).Error
	return
}

// validatePipeline 落库前统一校验
func (s *PipelineService) validatePipeline(p *workflow.Pipeline) error {
	if p.Name == "" {
		return errors.New("流水线名称不能为空")
	}
	if err := s.normalizeParamSchema(p); err != nil {
		return err
	}
	switch p.TriggerType {
	case workflow.TriggerManual, workflow.TriggerSchedule, workflow.TriggerWebhook:
	case "":
		p.TriggerType = workflow.TriggerManual
	default:
		return fmt.Errorf("triggerType 必须为 %s/%s/%s", workflow.TriggerManual, workflow.TriggerSchedule, workflow.TriggerWebhook)
	}
	nameSeen := map[string]bool{}
	for i := range p.Stages {
		st := &p.Stages[i]
		if st.Name == "" {
			return fmt.Errorf("第 %d 个阶段名称不能为空", i+1)
		}
		if nameSeen[st.Name] {
			return fmt.Errorf("阶段名称重复: %s", st.Name)
		}
		nameSeen[st.Name] = true
		if err := validateSteps(st.Steps); err != nil {
			return fmt.Errorf("阶段 %s: %w", st.Name, err)
		}
	}
	return nil
}

func validateSteps(steps []workflow.PipelineStep) error {
	for i := range steps {
		sp := &steps[i]
		if sp.Name == "" {
			return fmt.Errorf("第 %d 个步骤名称不能为空", i+1)
		}
		switch sp.Type {
		case workflow.StepTypeHTTP, workflow.StepTypeShell:
		default:
			return fmt.Errorf("步骤 %s 的 type 必须为 %s 或 %s", sp.Name, workflow.StepTypeHTTP, workflow.StepTypeShell)
		}
		if len(sp.Config) > 0 && !json.Valid(sp.Config) {
			return fmt.Errorf("步骤 %s 的 config 必须是合法 JSON", sp.Name)
		}
	}
	return nil
}

// normalizeParamSchema 保留为占位: ParamSchema 由 API 层绑定 CreatePipelineReq 时
// 已是 datatypes.JSON, Service 不重复序列化。此方法预留给未来需要二次校验时使用。
func (s *PipelineService) normalizeParamSchema(p *workflow.Pipeline) error {
	if len(p.ParamSchema) > 0 && !json.Valid(p.ParamSchema) {
		return errors.New("paramSchema 必须是合法 JSON")
	}
	return nil
}

// mustJSON 为 datatypes.JSON 构造帮助(供 BuildService 复用)
func mustJSON(v interface{}) datatypes.JSON {
	b, _ := json.Marshal(v)
	return b
}
