package task

import (
	"math"
	"time"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/vector"
)

// deployRenderer：Worker Pool 視覺（無狀態、依時間與進度動態）。
type deployRenderer struct {
	totalSec     int64
	remainingSec int64
	rewardK      int64
}

func NewDeployRenderer() Renderer { return &deployRenderer{} }

func (r *deployRenderer) Sync(totalSec, remainingSec int64, rewardK int64) {
	r.totalSec = totalSec
	r.remainingSec = remainingSec
	r.rewardK = rewardK
}
func (r *deployRenderer) Update(dtMs time.Duration) {}

func (r *deployRenderer) Draw(dst *ebiten.Image, x, y, w, h int, theme Theme) {
	if w <= 0 || h <= 0 {
		return
	}
	pad := 10
	// 進度（0..1）
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

	// 左：佇列槽
	qx := x + pad
	qy := y + pad
	qw := 18
	qh := h - 2*pad
	vector.StrokeRect(dst, float32(qx), float32(qy), float32(qw), float32(qh), 1, theme.Grid, true)
	// 佇列方塊數量隨 (1-pct)
	qn := int(3 + 9*(1-pct))
	if qn < 0 {
		qn = 0
	}
	gap := 4
	box := 6
	for i := 0; i < qn; i++ {
		by := float32(qy + 2 + i*(box+gap))
		if int(by)+box > qy+qh-2 {
			break
		}
		vector.DrawFilledRect(dst, float32(qx+3), by, float32(qw-6), float32(box), theme.Text, true)
	}

	// 中：Worker Pool 區
	wx := x + pad + qw + 6
	wy := y + pad
	ww := w - 2*pad - qw - 6 - qw
	wh := h - 2*pad
	if ww < 40 {
		ww = 40
	}
	vector.StrokeRect(dst, float32(wx), float32(wy), float32(ww), float32(wh), 1, theme.Grid, true)

	// worker 欄數（寬度決定）
	cols := int(math.Max(3, math.Min(6, float64(ww/90))))
	colW := ww / cols
	now := time.Now()
	baseT := float32((now.UnixNano() % 1_000_000_000)) / 1_000_000_000
	for c := 0; c < cols; c++ {
		// 每欄一個 worker 進度條
		cx := wx + c*colW
		// 欄框
		vector.StrokeRect(dst, float32(cx+4), float32(wy+4), float32(colW-8), float32(wh-8), 1, theme.Grid, true)
		// 進度（加入位移讓不同欄不同步）
		phase := float32(c) * 0.17
		prog := float32(math.Mod(float64(baseT+phase), 1.0))
		// 讓整體不小於任務實際進度（視覺上逐步接近完成）
		if prog < pct*0.85 {
			prog = pct * 0.85
		}
		bw := int(float32(colW-16) * prog)
		if bw < 0 {
			bw = 0
		}
		vector.DrawFilledRect(dst, float32(cx+8), float32(wy+8), float32(bw), 10, theme.Good, true)
		// 工作列下方流動小方塊（假裝任務輸入/輸出）
		dots := 3
		for i := 0; i < dots; i++ {
			dx := float32(cx + 8 + ((i*18)+int(baseT*float32(colW-24)))%(colW-24))
			dy := float32(wy + 24)
			vector.DrawFilledRect(dst, dx, dy, 6, 4, theme.Accent, true)
		}
	}

	// 右：完成槽，堆高隨進度
	fx := x + w - pad - qw
	fy := y + pad
	fh := h - 2*pad
	vector.StrokeRect(dst, float32(fx), float32(fy), float32(qw), float32(fh), 1, theme.Grid, true)
	fillH := int(float32(fh-4) * pct)
	if fillH > 0 {
		vector.DrawFilledRect(dst, float32(fx+2), float32(fy+fh-2-fillH), float32(qw-4), float32(fillH), theme.Good, true)
	}
}
