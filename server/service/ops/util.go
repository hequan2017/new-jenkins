package ops

import "time"

// todayStart 返回今天本地 0 点的 UTC 时间, 用于"今日"统计。
func todayStart() time.Time {
	now := time.Now()
	return time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
}
