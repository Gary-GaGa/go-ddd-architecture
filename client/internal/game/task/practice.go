package task

import (
	"math/rand"
	"time"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/vector"
)

// practiceRenderer 呈現 goroutine × channel 的簡化視覺。
type practiceRenderer struct {
	totalMs int
	leftMs  int
	// 視覺狀態
	spawnMs int
	msgs    []Rect // 訊息小方塊（在主通道上移動）
	// 群集匯流節奏
	burstTickerMs int
}

func NewPracticeRenderer() Renderer { return &practiceRenderer{} }

func (r *practiceRenderer) Sync(totalSec, remainingSec int64, _ int64) {
	r.totalMs = int(totalSec * 1000)
	r.leftMs = int(remainingSec * 1000)
	if r.leftMs < 0 {
		r.leftMs = 0
	}
	if r.totalMs <= 0 {
		r.totalMs = 1
	}
}

func (r *practiceRenderer) Update(dt time.Duration) {
	// 基礎密度提升：更頻繁地生成
	r.spawnMs -= int(dt / time.Millisecond)
	if r.spawnMs <= 0 {
		r.spawnMs = 220 + rand.Intn(160)
		// 生成單一訊息
		r.msgs = append(r.msgs, Rect{X: -1, Y: -1, W: 6, H: 6}) // 位置會在 Draw 前對齊到起點
	}
	// 每 ~2 秒做一次大量匯流
	r.burstTickerMs -= int(dt / time.Millisecond)
	if r.burstTickerMs <= 0 {
		r.burstTickerMs = 2000 + rand.Intn(400)
		// 一次性塞入多顆訊息，造成節點周邊密集
		n := 4 + rand.Intn(4)
		for i := 0; i < n; i++ {
			r.msgs = append(r.msgs, Rect{X: -1, Y: -1, W: 6, H: 6})
		}
	}
	// 推進既有訊息
	for i := range r.msgs {
		r.msgs[i].X += 1.0 // 稍微加速，純視覺
	}
	// 清掉超出範圍的（寬度未知，留給 Draw 做裁切）
	if len(r.msgs) > 48 { // 上限保護
		r.msgs = r.msgs[len(r.msgs)-48:]
	}
}

func (r *practiceRenderer) Draw(dst *ebiten.Image, x, y, w, h int, theme Theme) {
	if w <= 0 || h <= 0 {
		return
	}
	// 背景通道（2 條）
	pad := 10
	laneY1 := float32(y + pad + 14)
	laneY2 := float32(y + h - pad - 18)
	vector.StrokeLine(dst, float32(x+pad), laneY1, float32(x+w-pad), laneY1, 1, theme.Grid, true)
	vector.StrokeLine(dst, float32(x+pad), laneY2, float32(x+w-pad), laneY2, 1, theme.Grid, true)
	// select 節點（中間一個，分支到下通道）
	nodeX := float32(x + w/2)
	vector.StrokeLine(dst, nodeX, laneY1, nodeX, laneY2, 1, theme.Accent, true)
	// 節點徽章：阻塞/非阻塞（視覺偵測：節點附近訊息多且下通道擁擠時為 Warn）
	nearCount := 0
	lowerBusy := 0
	for _, m := range r.msgs {
		if absf(m.X-nodeX) < 8 {
			nearCount++
		}
		if m.Y > laneY2-6 {
			lowerBusy++
		}
	}
	badgeCol := theme.Good
	if nearCount >= 3 && lowerBusy >= 2 {
		badgeCol = theme.Warn
	}
	// 小徽章畫在節點上方
	vector.DrawFilledRect(dst, nodeX-3, laneY1-10, 6, 6, badgeCol, true)
	// 訊息小方塊（只在上通道生成並通過節點往下漂）
	progress := 1 - float32(r.leftMs)/float32(r.totalMs)
	if progress < 0 {
		progress = 0
	}
	if progress > 1 {
		progress = 1
	}
	// 根據剩餘時間密度生成/位置對齊
	if r.spawnMs <= 30 {
		// 在通道起點生成
		r.msgs = append(r.msgs, Rect{X: float32(x + pad), Y: laneY1 - 3, W: 6, H: 6})
	}
	for i := range r.msgs {
		m := &r.msgs[i]
		if m.X < 0 { // 初始對齊到起點
			m.X = float32(x + pad)
			m.Y = laneY1 - 3
		}
		// 當經過節點時（x > nodeX）緩慢往下靠攏
		if float32(m.X) > nodeX-6 {
			m.Y += 0.5
			if m.Y > laneY2-3 {
				m.Y = laneY2 - 3
			}
		}
		// 繪製（視窗內才畫）
		if float32(m.X) >= float32(x+pad) && float32(m.X) <= float32(x+w-pad) {
			vector.DrawFilledRect(dst, m.X, m.Y, m.W, m.H, theme.Key, true)
		}
	}
	// 進度匯流（底線）
	px := float32(x + pad)
	pw := float32(w - pad*2)
	vector.StrokeLine(dst, px, float32(y+h-8), px+pw*progress, float32(y+h-8), 2, theme.Accent, true)
}

// absf returns absolute value for float32
func absf(v float32) float32 {
	if v < 0 {
		return -v
	}
	return v
}
