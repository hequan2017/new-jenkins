package workflow

import (
	"context"
	"fmt"

	"github.com/flipped-aurora/gin-vue-admin/server/global"
	"github.com/flipped-aurora/gin-vue-admin/server/model/workflow"
	"github.com/flipped-aurora/gin-vue-admin/server/utils/datascope"
	"github.com/flipped-aurora/gin-vue-admin/server/utils/logger"
	"github.com/robfig/cron/v3"
)

// WorkflowScheduleService 流水线定时调度(复用 global.GVA_TIMER, 仿 sys_timed_task 范式)
type WorkflowScheduleService struct{}

// pipelineCronName 一流水线一 cronName: robfig/cron 无单任务 pause,
// 启停语义 = Clear(cronName) + 按 DB 重加, 状态以 DB enabled 为准。
func pipelineCronName(id uint) string { return fmt.Sprintf("workflow/pipeline/%d", id) }

// ValidateSpec 服务端校验 cron 表达式(含 @daily/@hourly 等描述符)
func ValidateSpec(spec string, withSeconds bool) error {
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

// SchedulePipeline 幂等调度: 先 Clear 再按 enabled+spec 重新注册
// 回调内用系统上下文(datascope.WithSystem)触发构建, trigger=schedule, triggerBy=0(系统)
func (sch *WorkflowScheduleService) SchedulePipeline(p workflow.Pipeline) error {
	name := pipelineCronName(p.ID)
	global.GVA_Timer.Clear(name)
	if !p.Enabled || p.TriggerType != workflow.TriggerSchedule || p.Spec == "" {
		return nil
	}
	id := p.ID
	fn := func() {
		// 系统上下文触发, 避免 datascope 旁路告警
		sysCtx := datascope.WithSystem(context.Background())
		_, err := (&BuildService{}).TriggerBuild(sysCtx, id, nil, workflow.TriggerSchedule, 0)
		if err != nil {
			logger.Bg().Mod("workflow").Err(err).Error(fmt.Sprintf("定时触发流水线 %d 失败", id))
		}
	}
	var err error
	if p.WithSeconds {
		_, err = global.GVA_Timer.AddTaskByFuncWithSecond(name, p.Spec, fn, p.Name)
	} else {
		_, err = global.GVA_Timer.AddTaskByFunc(name, p.Spec, fn, p.Name)
	}
	return err
}

// LoadAll 启动/重载后按 DB 恢复调度(仅 schedule 类型 + enabled + 有 spec 的注册; 幂等)
func (sch *WorkflowScheduleService) LoadAll(ctx context.Context) error {
	var pipelines []workflow.Pipeline
	if err := global.GVA_DB.WithContext(ctx).
		Where("trigger_type = ?", workflow.TriggerSchedule).Find(&pipelines).Error; err != nil {
		return err
	}
	for i := range pipelines {
		if err := sch.SchedulePipeline(pipelines[i]); err != nil {
			logger.Bg().Mod("workflow").Err(err).Error("流水线恢复调度失败: " + pipelines[i].Name)
		}
	}
	logger.Bg().Mod("workflow").Info(fmt.Sprintf("定时流水线加载完成, 共 %d 条", len(pipelines)))
	return nil
}
