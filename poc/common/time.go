package common

// Nano2Milli 纳秒转换成毫秒
func Nano2Milli(ns int64) int {
	return int(ns / 1e6)
}
