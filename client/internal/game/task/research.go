package task

import (
	"math"
	"time"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/vector"
)

// researchRenderer：Profiler/GC 波形 + 管線進度（無狀態、依時間/進度）。
type researchRenderer struct {
	totalSec     int64
	remainingSec int64
	rewardK      int64
}

func NewResearchRenderer() Renderer { return &researchRenderer{} }

func (r *researchRenderer) Sync(totalSec, remainingSec int64, rewardK int64) {
	r.totalSec = totalSec
	r.remainingSec = remainingSec
	r.rewardK = rewardK
}
func (r *researchRenderer) Update(dt time.Duration) {}

func (r *researchRenderer) Draw(dst *ebiten.Image, x, y, w, h int, theme Theme) {
	if w <= 0 || h <= 0 {
		return
	}
	pad := 10
	pct := float32(0)
	if r.totalSec > 0 {
		done := r.totalSec - r.remainingSec
		if done < 0 {
			done = 0
		}
		if done > r.totalSec {
			done = r.totalSec
		}
		pct = float32(done) / float32(r.totalSec)
	}
	// 上半：管線
	ph := (h - 2*pad) / 2
	stages := 5
	cell := (w - 2*pad) / stages
	for i := 0; i < stages; i++ {
		bx := float32(x + pad + i*cell + 6)
		by := float32(y + pad + 6)
		vector.StrokeRect(dst, bx, by, float32(cell-12), 14, 1, theme.Grid, true)
		// 已完成比例（將任務進度映射到階段亮度/填充）
		seg := float32(i+1) / float32(stages)
		if pct >= seg {
			vector.DrawFilledRect(dst, bx+1, by+1, float32(cell-14), 12, theme.Good, true)
		} else if pct > seg-1.0/float32(stages) {
			// 當前進行中的階段：部分填充
			local := (pct - (seg - 1.0/float32(stages))) * float32(stages)
			ww := int(float32(cell-14) * local)
			if ww > 0 {
				vector.DrawFilledRect(dst, bx+1, by+1, float32(ww), 12, theme.Accent, true)
			}
		}
	}

	// 下半：GC 掃描波與樣本方塊
	baseY := float32(y + pad + ph + 8)
	width := w - 2*pad
	now := time.Now()
	t := float32((now.UnixNano() % 1_000_000_000)) / 1_000_000_000
	// 波形（2 條：主波與掃描線）
	for i := 0; i < width; i++ {
		x0 := float32(x + pad + i)
		// 主波：慢速正弦，幅度 8
		y0 := baseY + 4*float32(math.Sin(float64(i)*0.05+float64(t)*2.2))
		vector.DrawFilledRect(dst, x0, y0, 1, 8, theme.Text, true)
	}
	// 掃描線：每 ~1.8s 從左到右掃過一次
	sweep := int(float32(width) * float32(math.Mod(float64(t*0.55), 1.0)))
	sx := float32(x + pad + sweep)
	vector.DrawFilledRect(dst, sx, baseY-4, 2, 24, theme.Warn, true)

	// 樣本方塊：依進度增加密度
	step := int(math.Max(10, 28-18*float64(pct)))
	if step < 6 {
		step = 6
	}
	for i := 0; i < width; i += step {
		hh := 6 + (i/step)%5
		vector.DrawFilledRect(dst, float32(x+pad+i), baseY+20, 6, float32(hh), theme.Accent, true)
	}
}
