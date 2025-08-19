package task

import (
	"image/color"
)

// TaskKind 描述任務類型。
// Practice: 練習；Research: 研究；Deploy: 部署
// 預留給未來擴充對應不同動畫風格。
type TaskKind int

const (
	Practice TaskKind = iota
	Research
	Deploy
)

// EntityKind 場景中的物件種類。
type EntityKind int

const (
	Key EntityKind = iota
	Obstacle
	Bonus
)

// Rect 與 Vel 為簡單的幾何結構。
type Rect struct{ X, Y, W, H float32 }

type Vel struct{ VX, VY float32 }

// Entity 場景實體。
type Entity struct {
	Rect  Rect
	Vel   Vel
	Kind  EntityKind
	Alive bool
	Col   color.RGBA
}

// Scene 表示一個正在執行中的任務動畫場景。
type Scene struct {
	Kind       TaskKind
	Obstacles  []Entity
	Bonuses    []Entity
	Progress   float32 // 0..1
	TimeLeftMs int
	RewardK    int64

	// internal
	width, height int
	startX, endX  float32
	spawnTickerMs int
	totalMs       int
	Key           Entity
}

// Theme 提供繪製時使用的色彩。
type Theme struct {
	Grid, Key, Goal, Bonus, Text, Accent color.RGBA
	Good, Warn                           color.RGBA
}
