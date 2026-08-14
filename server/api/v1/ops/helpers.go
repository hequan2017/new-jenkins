package ops

import "strconv"

// parseUint 解析 query 参数为 uint, 失败返回 0。
func parseUint(s string) uint {
	if s == "" {
		return 0
	}
	v, err := strconv.ParseUint(s, 10, 64)
	if err != nil {
		return 0
	}
	return uint(v)
}

// parseInt 解析 query 参数为 int, 失败返回 0。
func parseInt(s string) int {
	if s == "" {
		return 0
	}
	v, err := strconv.Atoi(s)
	if err != nil {
		return 0
	}
	return v
}
