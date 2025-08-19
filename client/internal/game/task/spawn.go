package task

import (
	"math/rand"
	"time"
)

// Spawn 在 400~700ms 隨機生成視覺小物件。
func (s *Scene) Spawn(now time.Time) {
	if s.Progress >= 1 {
		return
	}
	if s.spawnTickerMs <= 0 {
		s.spawnTickerMs = 400 + rand.Intn(300)
	}
	s.spawnTickerMs -= 16 // 約略以 60fps 遞減
	if s.spawnTickerMs > 0 {
		return
	}
	s.spawnTickerMs = 400 + rand.Intn(300)
	if rand.Float32() < 0.5 {
		// spawn obstacle
		y := float32(rand.Intn(s.height-24) + 12)
		s.Obstacles = append(s.Obstacles, Entity{Rect: Rect{X: float32(rand.Intn(40)) + s.startX + 30, Y: y, W: 6, H: 6}, Kind: Obstacle, Alive: true})
	} else {
		// spawn bonus
		y := float32(rand.Intn(s.height-24) + 12)
		s.Bonuses = append(s.Bonuses, Entity{Rect: Rect{X: float32(rand.Intn(40)) + s.startX + 30, Y: y, W: 4, H: 4}, Kind: Bonus, Alive: true})
	}
}
