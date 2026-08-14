package ops

import (
	"context"

	"github.com/flipped-aurora/gin-vue-admin/server/global"
	"github.com/flipped-aurora/gin-vue-admin/server/model/ops"
	"github.com/flipped-aurora/gin-vue-admin/server/model/workflow"
)

// DashboardService 运维大盘聚合数据查询。
type DashboardService struct{}

// DashboardData 大盘聚合返回结构
type DashboardData struct {
	AssetTotal    int64            `json:"assetTotal"`
	AssetOnline   int64            `json:"assetOnline"`
	CredentialCnt int64            `json:"credentialCnt"`
	TicketStat    map[string]int64 `json:"ticketStat"`
	BuildRecent   map[string]int64 `json:"buildRecent"` // 近 20 次构建状态计数
	InspectStat   map[string]int64 `json:"inspectStat"` // 各巡检任务最近状态
	AlertToday    int64            `json:"alertToday"`  // 今日巡检异常数
	AuditToday    int64            `json:"auditToday"`  // 今日审计数
	AlertActive   int64            `json:"alertActive"` // 未处理告警数
}

// GetDashboard 聚合各维度统计
func (s *DashboardService) GetDashboard(ctx context.Context) (DashboardData, error) {
	d := DashboardData{
		TicketStat:  map[string]int64{},
		BuildRecent: map[string]int64{},
		InspectStat: map[string]int64{},
	}
	db := global.GVA_DB.WithContext(ctx)

	db.Model(&ops.Asset{}).Count(&d.AssetTotal)
	db.Model(&ops.Asset{}).Where("status = ?", ops.AssetStatusOnline).Count(&d.AssetOnline)
	db.Model(&ops.Credential{}).Count(&d.CredentialCnt)

	// 工单状态分组
	var ticketRows []struct {
		Status string
		Cnt    int64
	}
	db.Model(&ops.Ticket{}).Select("status, count(*) as cnt").Group("status").Scan(&ticketRows)
	for _, r := range ticketRows {
		d.TicketStat[r.Status] = r.Cnt
	}

	// 近 20 次构建状态分布
	var builds []workflow.PipelineBuild
	db.Order("id DESC").Limit(20).Find(&builds)
	for _, b := range builds {
		d.BuildRecent[b.Status]++
	}

	// 巡检: 每个启用任务最近一次结果状态
	var tasks []ops.InspectTask
	db.Where("enabled = ?", true).Find(&tasks)
	for _, t := range tasks {
		var r ops.InspectResult
		if err := db.Where("task_id = ?", t.ID).Order("id DESC").First(&r).Error; err == nil {
			d.InspectStat[r.Status]++
		}
	}

	// 今日巡检异常 + 审计数
	db.Model(&ops.InspectResult{}).
		Where("status = ?", ops.InspectStatusAlert).
		Where("created_at >= ?", todayStart()).Count(&d.AlertToday)
	db.Model(&ops.AuditRecord{}).Where("created_at >= ?", todayStart()).Count(&d.AuditToday)
	d.AlertActive = (&AlertService{}).CountActive(ctx)

	return d, nil
}
