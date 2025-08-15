package clock

import "time"

// SystemClock 回傳系統當前時間。

type SystemClock struct{}

func (SystemClock) Now() time.Time { return time.Now() }
