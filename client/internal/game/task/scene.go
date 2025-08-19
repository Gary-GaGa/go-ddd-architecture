package task

import (
	"time"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/vector"
)

// NewScene 建立一個基本模式（basic）場景。
func NewScene(kind TaskKind, w, h int, rewardK int64, duration time.Duration) *Scene {
	s := &Scene{
		Kind:       kind,
		Progress:   0,
		TimeLeftMs: int(duration / time.Millisecond),
		RewardK:    rewardK,
		width:      w,
		height:     h,
		startX:     16,
		endX:       float32(w - 32),
	}
	s.totalMs = s.TimeLeftMs
	s.Key = Entity{
		Rect:  Rect{X: s.startX, Y: float32(h/2 - 6), W: 12, H: 12},
		Kind:  Key,
		Alive: true,
	}
	return s
}

// Update 根據 dt 推進進度，並移動鑰匙。
func (s *Scene) Update(dt time.Duration) {
	if s.Progress >= 1 {
		return
	}
	if s.TimeLeftMs > 0 {
		s.TimeLeftMs -= int(dt / time.Millisecond)
		if s.TimeLeftMs < 0 {
			s.TimeLeftMs = 0
		}
	}
	// 根據已流逝時間推進 progress（線性）
	if s.totalMs > 0 {
		elapsed := s.totalMs - s.TimeLeftMs
		if elapsed < 0 {
			elapsed = 0
		}
		if elapsed > s.totalMs {
			elapsed = s.totalMs
		}
		s.Progress = float32(elapsed) / float32(s.totalMs)
	}
	if s.Progress > 1 {
		s.Progress = 1
	}
	// 鑰匙位置：線性插值
	x := s.startX + (s.endX-s.startX)*s.Progress
	s.Key.Rect.X = x
	// 小事件生成（視覺用）
	s.Spawn(time.Now())
}

// SyncRemaining 與 VM 同步任務的總時長與剩餘時間（毫秒）。
// 若 totalSec<=0 則忽略；remainingSec 會被夾在 [0, totalSec]。
func (s *Scene) SyncRemaining(totalSec, remainingSec int64) {
	if totalSec <= 0 {
		return
	}
	totalMs := int(totalSec * 1000)
	remMs := int(remainingSec * 1000)
	if remMs < 0 {
		remMs = 0
	}
	if remMs > totalMs {
		remMs = totalMs
	}
	s.totalMs = totalMs
	s.TimeLeftMs = remMs
	// 依據同步後的時間重算進度
	elapsed := totalMs - remMs
	if elapsed < 0 {
		elapsed = 0
	}
	if elapsed > totalMs {
		elapsed = totalMs
	}
	if totalMs > 0 {
		s.Progress = float32(elapsed) / float32(totalMs)
	} else {
		s.Progress = 1
	}
	// 同步位置
	s.Key.Rect.X = s.startX + (s.endX-s.startX)*s.Progress
}

// Draw 繪製網格、Key、終點方塊；完成時顯示 +K 漂浮字。
func (s *Scene) Draw(dst *ebiten.Image, originX, originY int, theme Theme) {
	// 繪製背景網格
	for x := originX; x < originX+s.width; x += 12 {
		vector.StrokeLine(dst, float32(x), float32(originY), float32(x), float32(originY+s.height), 1, theme.Grid, true)
	}
	for y := originY; y < originY+s.height; y += 12 {
		vector.StrokeLine(dst, float32(originX), float32(y), float32(originX+s.width), float32(y), 1, theme.Grid, true)
	}
	// 目標方塊
	vector.StrokeRect(dst, float32(originX+s.width-24), float32(originY+s.height/2-12), 20, 20, 2, theme.Goal, true)
	// 鑰匙（用小方塊替代，與 HUD 風格一致）
	vector.DrawFilledRect(dst, float32(originX)+s.Key.Rect.X, float32(originY)+s.Key.Rect.Y, s.Key.Rect.W, s.Key.Rect.H, theme.Key, true)
	// 障礙與獎勵
	for i := range s.Obstacles {
		o := &s.Obstacles[i]
		if !o.Alive {
			continue
		}
		if i%2 == 0 {
			o.Rect.Y += 0.2
		} else {
			o.Rect.Y -= 0.2
		}
		vector.DrawFilledRect(dst, float32(originX)+o.Rect.X, float32(originY)+o.Rect.Y, o.Rect.W, o.Rect.H, theme.Accent, true)
	}
	for i := range s.Bonuses {
		b := &s.Bonuses[i]
		if !b.Alive {
			continue
		}
		vector.DrawFilledRect(dst, float32(originX)+b.Rect.X, float32(originY)+b.Rect.Y, b.Rect.W, b.Rect.H, theme.Bonus, true)
	}
	// 完成特效
	if s.Progress >= 1 {
		// 簡易星芒
		cx := float32(originX + s.width - 14)
		cy := float32(originY + s.height/2)
		vector.StrokeLine(dst, cx-8, cy, cx+8, cy, 2, theme.Accent, true)
		vector.StrokeLine(dst, cx, cy-8, cx, cy+8, 2, theme.Accent, true)
	}
}

// 色彩 RGBA（避免與 HUD 依賴循環）
// colorRGBA 已移除，改用 Theme 直接攜帶 color.RGBA
