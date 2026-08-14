package ops

import (
	"context"

	"github.com/flipped-aurora/gin-vue-admin/server/global"
	"github.com/flipped-aurora/gin-vue-admin/server/model/ops"
	"github.com/flipped-aurora/gin-vue-admin/server/model/workflow"
)

// ScheduleCenterService 调度任务中心: 聚合 workflow schedule 流水线与 ops 巡检/备份任务,
// 提供统一查看与手动触发入口。本身不新增表, 只读既有模块数据。
type ScheduleCenterService struct{}

// ScheduleItem 统一的调度任务视图
type ScheduleItem struct {
	Source     string `json:"source"`     // workflow | inspect | backup
	ID         uint   `json:"id"`
	Name       string `json:"name"`
	Spec       string `json:"spec"`
	Enabled    bool   `json:"enabled"`
	RefName    string `json:"refName"`    // 关联对象(如资产名/流水线说明)
	NextField  string `json:"nextField"`  // 预留: 下次执行时间(可后续扩展)
}

// GetList 聚合各来源调度任务
func (s *ScheduleCenterService) GetList(ctx context.Context, source string) (list []ScheduleItem, total int64, err error) {
	db := global.GVA_DB.WithContext(ctx)
	list = []ScheduleItem{}

	if source == "" || source == "workflow" {
		var pips []workflow.Pipeline
		db.Where("trigger_type = ?", workflow.TriggerSchedule).Find(&pips)
		for _, p := range pips {
			list = append(list, ScheduleItem{
				Source: "workflow", ID: p.ID, Name: p.Name, Spec: p.Spec, Enabled: p.Enabled,
				RefName: p.Description,
			})
		}
	}
	if source == "" || source == "inspect" {
		var tasks []ops.InspectTask
		db.Find(&tasks)
		for _, t := range tasks {
			list = append(list, ScheduleItem{
				Source: "inspect", ID: t.ID, Name: t.Name, Spec: t.Spec, Enabled: t.Enabled,
				RefName: "巡检",
			})
		}
	}
	if source == "" || source == "backup" {
		var tasks []ops.BackupTask
		db.Find(&tasks)
		for _, t := range tasks {
			list = append(list, ScheduleItem{
				Source: "backup", ID: t.ID, Name: t.Name, Spec: t.Spec, Enabled: t.Enabled,
				RefName: t.RemotePath,
			})
		}
	}
	total = int64(len(list))
	return
}
