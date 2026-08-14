// server/initialize/ops_schedule.go
package initialize

import (
	"github.com/flipped-aurora/gin-vue-admin/server/service"
)

// LoadOpsInspectSchedules 从 ops_inspect_tasks 恢复启用的巡检任务调度(幂等)。
// 必须在 RegisterTables(建表)之后调用。
func LoadOpsInspectSchedules() {
	// 巡检任务可能为空, 内部静默处理; 不阻断启动
	service.ServiceGroupApp.OpsServiceGroup.InspectService.LoadSchedules()
}
