// server/initialize/ops_schedule.go
package initialize

import (
	"github.com/flipped-aurora/gin-vue-admin/server/service"
)

// LoadOpsInspectSchedules 从 ops_inspect_tasks 恢复启用的巡检任务调度(幂等)。
func LoadOpsInspectSchedules() {
	service.ServiceGroupApp.OpsServiceGroup.InspectService.LoadSchedules()
}

// LoadOpsBackupSchedules 从 ops_backup_tasks 恢复启用的备份任务调度(幂等)。
func LoadOpsBackupSchedules() {
	service.ServiceGroupApp.OpsServiceGroup.BackupService.LoadSchedules()
}
