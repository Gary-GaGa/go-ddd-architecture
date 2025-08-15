package gametime

import "time"

// Timestamps 保存關閉時的牆鐘時間與單調時間代理值。
type Timestamps struct {
	WallClockAtClose        time.Time
	ElapsedMonotonicSeconds int64
}
