package ui

import (
	"fmt"
	"image/color"
	"math"
	"time"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/text"
	"github.com/hajimehoshi/ebiten/v2/vector"
	"golang.org/x/image/font"
)

// DrawHUD draws the main HUD: resources, level, current task timer, etc.
func DrawHUD(screen *ebiten.Image, face font.Face, vm VM, errMsg string, inFlight bool) {
	// background
	screen.Fill(Theme.Bg)

	sw, sh := screen.Size()
	pad := Theme.Pad8

	// layout widths
	leftW := 280
	rightW := 260
	// center width fills remaining space with padding
	centerW := sw - leftW - rightW - pad*6
	if centerW < 240 {
		centerW = 240
	}

	// Left: Status card
	leftX, leftY := pad*2, pad*2
	leftH := 220
	drawCard(screen, leftX, leftY, leftW, leftH)
	tx, ty := leftX+pad, leftY+pad+14
	text.Draw(screen, "Status", face, tx, ty, Theme.TextMain)
	ty += 8
	vector.StrokeLine(screen, float32(tx), float32(ty), float32(leftX+leftW-pad), float32(ty), 1, Theme.CardBorder, true)
	ty += 12
	text.Draw(screen, fmt.Sprintf("Knowledge: %d", vm.Knowledge), face, tx, ty, Theme.TextMain)
	ty += 18
	text.Draw(screen, fmt.Sprintf("Research:  %d", vm.Research), face, tx, ty, Theme.TextMain)
	ty += 18
	text.Draw(screen, fmt.Sprintf("Level: %d", vm.Level), face, tx, ty, Theme.TextMain)
	costStr := fmt.Sprintf("Cost R %d", vm.NextUpgradeCost)
	cb := text.BoundString(face, costStr)
	badgeW := 6*2 + cb.Dx()
	drawBadge(screen, leftX+leftW-pad-badgeW, ty-12, costStr, face, canAfford(vm))
	ty += 22
	text.Draw(screen, fmt.Sprintf("Rates K/R per min: %d / %d", vm.KnowledgePerMin, vm.ResearchPerMin), face, tx, ty, Theme.TextSub)
	ty += 18
	if vm.EstimatedSuccess > 0 {
		text.Draw(screen, fmt.Sprintf("Est. Success: %.0f%%", vm.EstimatedSuccess*100), face, tx, ty, Theme.TextSub)
		ty += 22
	}

	// Networking / Error in left card bottom area
	nx := leftX + pad
	ny := leftY + leftH - pad - 36
	if inFlight {
		text.Draw(screen, "Networking...", face, nx, ny, Theme.Good)
		ny += 18
	}
	if errMsg != "" {
		text.Draw(screen, "Error: "+errMsg, face, nx, ny, Theme.Error)
		ny += 18
	}

	// Center: Task (top) + Shiba (bottom)
	centerX := leftX + leftW + pad*2
	centerY := leftY
	taskH := 140
	drawCard(screen, centerX, centerY, centerW, taskH)
	ttx, tty := centerX+pad, centerY+pad+12
	if vm.CurrentTask != nil {
		title := "Task: " + vm.CurrentTask.Type
		if vm.CurrentTask.Language != "" {
			title += " [" + vm.CurrentTask.Language + "]"
		}
		text.Draw(screen, title, face, ttx, tty, Theme.TextMain)
		tty += 16
		if vm.CurrentTask.Language != "" {
			startLabel := "Started: " + vm.CurrentTask.Language
			if vm.CurrentLanguage != "" && vm.CurrentLanguage != vm.CurrentTask.Language {
				startLabel += " (now: " + vm.CurrentLanguage + ")"
			}
			text.Draw(screen, startLabel, face, ttx, tty, Theme.Good)
			tty += 16
		} else {
			tty += 2
		}
		if vm.CurrentTask.BaseReward > 0 {
			text.Draw(screen, fmt.Sprintf("Base +%d K", vm.CurrentTask.BaseReward), face, ttx, tty, Theme.TextSub)
			tty += 16
		}
		if vm.CurrentTask.DurationSeconds > 0 {
			text.Draw(screen, fmt.Sprintf("Duration: %ds", vm.CurrentTask.DurationSeconds), face, ttx, tty, Theme.TextSub)
		}
		// radial timer on the right side of the task card
		remain := vm.CurrentTask.RemainingSeconds
		cx := centerX + centerW - pad - 40
		cy := centerY + taskH/2
		ratio := float32(remain%60) / 60.0
		if remain <= 10 {
			drawArcRing(screen, cx, cy, 20, ratio, Theme.Warn, Theme.Warn)
		} else {
			drawArcRing(screen, cx, cy, 20, ratio, Theme.Good, Theme.Warn)
		}
		label := fmt.Sprintf("%02d:%02d", remain/60, remain%60)
		b := text.BoundString(face, label)
		lx := cx - b.Dx()/2
		ly := cy + 5
		text.Draw(screen, label, face, lx, ly, Theme.TextSub)
	} else {
		text.Draw(screen, "No active task", face, ttx, tty, Theme.TextSub)
	}

	// Center: Shiba / visual area under the task
	shibaY := centerY + taskH + pad
	shibaH := 220
	drawCard(screen, centerX, shibaY, centerW, shibaH)
	// draw pixel Shiba centered in the shiba card
	// call existing DrawEvolvingAI to render the shiba into the card area
	DrawEvolvingAI(screen, face, centerX+8, shibaY+8, centerW-16, shibaH-16, 0, vm.CurrentLanguage, 0)

	// Right: Languages panel
	rightX := centerX + centerW + pad*2
	rightY := leftY
	// compute dynamic height for languages (fit content)
	rows := len(vm.Languages)
	langRowH := 24
	langH := 28 + max(0, rows*langRowH) + pad
	if langH < 120 {
		langH = 120
	}
	drawCard(screen, rightX, rightY, rightW, langH)
	lx, ly := rightX+pad, rightY+pad+12
	text.Draw(screen, "Languages", face, lx, ly, Theme.TextMain)
	ly += 16
	order := []string{"go", "py", "js"}
	seen := map[string]bool{}
	row := func(code string, s LangHUD) {
		col := Theme.TextSub
		if code == vm.CurrentLanguage {
			col = Theme.Good
		}
		label := fmt.Sprintf("%s  K:%d  R:%d  Lv:%d", code, s.Knowledge, s.Research, s.Level)
		text.Draw(screen, label, face, lx, ly, col)
		ly += langRowH
	}
	for _, k := range order {
		if s, ok := vm.Languages[k]; ok {
			row(k, s)
			seen[k] = true
		}
	}
	for k, s := range vm.Languages {
		if seen[k] {
			continue
		}
		row(k, s)
	}

	// Controls hint (bottom center)
	controls := "Controls: P Practice | U Upgrade | C Claim | 1 Go | 2 Python | 3 JavaScript"
	bw := text.BoundString(face, controls).Dx()
	cx := (sw - bw) / 2
	text.Draw(screen, controls, face, cx, sh-pad, Theme.TextSub)
}

// VM 是繪圖所需的最小視圖（由 State 快照轉換而來）。
type VM struct {
	Knowledge        int64
	Research         int64
	Level            int
	NextUpgradeCost  int64
	KnowledgePerMin  int64
	ResearchPerMin   int64
	CurrentTask      *VMTask
	EstimatedSuccess float64
	// Multi-language 額外資訊（用於左側語言卡片）
	CurrentLanguage string
	Languages       map[string]LangHUD
}

type VMTask struct {
	Type             string
	RemainingSeconds int64
	Language         string
	DurationSeconds  int64
	BaseReward       int64
}

// LangHUD 是 HUD 顯示用的語言統計摘要。
type LangHUD struct {
	Code      string
	Knowledge int64
	Research  int64
	Level     int
}

// --- styled helpers ---

func drawCard(dst *ebiten.Image, x, y, w, h int) {
	// 陰影（輕微）
	shadow := color.RGBA{0, 0, 0, 36}
	drawRoundedFilledRect(dst, x+2, y+3, w, h, Theme.Radius8, shadow)
	// 本體
	drawRoundedFilledRect(dst, x, y, w, h, Theme.Radius8, Theme.CardBg)
	// 外框（圍繞圓角邊）
	drawRoundedRectOutline(dst, x, y, w, h, Theme.Radius8, Theme.CardBorder, 1)
}

func drawBadge(dst *ebiten.Image, x, y int, textStr string, face font.Face, ok bool) {
	padX, padY := 6, 4
	b := text.BoundString(face, textStr)
	w := padX*2 + b.Dx()
	h := padY*2 + 14
	// 背景色：可負擔用 Good，不可負擔用 Warn，並做小幅脈動（提示不足）
	var bg color.RGBA
	if ok {
		bg = Theme.Good
	} else {
		// 以 1.2 秒週期讓 alpha 在 180~255 摆動
		period := float64(1200 * time.Millisecond)
		t := float64(time.Now().UnixNano()%int64(period)) / period
		a := uint8(180 + 75*0.5*(1+math.Sin(2*math.Pi*t)))
		bg = color.RGBA{R: Theme.Warn.R, G: Theme.Warn.G, B: Theme.Warn.B, A: a}
	}
	drawRoundedFilledRect(dst, x, y, w, h, 6, Theme.CardBorder)
	drawRoundedFilledRect(dst, x+1, y+1, w-2, h-2, 5, bg)
	text.Draw(dst, textStr, face, x+padX, y+padY+12, color.Black)
}

func drawArcRing(dst *ebiten.Image, cx, cy, r int, ratio float32, colStart, colEnd color.RGBA) {
	// 背景環
	vector.StrokeCircle(dst, float32(cx), float32(cy), float32(r), 2, Theme.CardBorder, true)
	// 前景弧：以多段近似連續弧形
	if ratio <= 0 {
		return
	}
	if ratio > 1 {
		ratio = 1
	}
	seg := 120
	end := int(float32(seg) * ratio)
	var prevX, prevY float32
	for i := 0; i <= end; i++ {
		ang := -math.Pi/2 + 2*math.Pi*float64(i)/float64(seg)
		x := float32(cx) + float32(r)*float32(math.Cos(ang))
		y := float32(cy) + float32(r)*float32(math.Sin(ang))
		if i > 0 {
			t := float32(i) / float32(max(1, end))
			col := lerpRGBA(colStart, colEnd, t)
			vector.StrokeLine(dst, prevX, prevY, x, y, 3, col, true)
		}
		prevX, prevY = x, y
	}
}

func lerpRGBA(a, b color.RGBA, t float32) color.RGBA {
	if t < 0 {
		t = 0
	} else if t > 1 {
		t = 1
	}
	lerp := func(x, y uint8) uint8 { return uint8(float32(x) + (float32(y)-float32(x))*t + 0.5) }
	return color.RGBA{
		R: lerp(a.R, b.R),
		G: lerp(a.G, b.G),
		B: lerp(a.B, b.B),
		A: lerp(a.A, b.A),
	}
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}

// --- rounded rect helpers ---

func drawRoundedFilledRect(dst *ebiten.Image, x, y, w, h int, radius float32, col color.Color) {
	if w <= 0 || h <= 0 {
		return
	}
	r := radius
	if r < 0 {
		r = 0
	}
	if float32(w) < 2*r {
		r = float32(w) / 2
	}
	if float32(h) < 2*r {
		r = float32(h) / 2
	}
	// 中央長方形
	vector.DrawFilledRect(dst, float32(x)+r, float32(y), float32(w)-2*r, float32(h), col, true)
	// 左右直條
	vector.DrawFilledRect(dst, float32(x), float32(y)+r, r, float32(h)-2*r, col, true)
	vector.DrawFilledRect(dst, float32(x)+float32(w)-r, float32(y)+r, r, float32(h)-2*r, col, true)
	// 四角圓
	vector.DrawFilledCircle(dst, float32(x)+r, float32(y)+r, r, col, true)
	vector.DrawFilledCircle(dst, float32(x)+float32(w)-r, float32(y)+r, r, col, true)
	vector.DrawFilledCircle(dst, float32(x)+r, float32(y)+float32(h)-r, r, col, true)
	vector.DrawFilledCircle(dst, float32(x)+float32(w)-r, float32(y)+float32(h)-r, r, col, true)
}

func drawRoundedRectOutline(dst *ebiten.Image, x, y, w, h int, radius float32, col color.Color, width float32) {
	if w <= 0 || h <= 0 {
		return
	}
	r := radius
	if r < 0 {
		r = 0
	}
	if float32(w) < 2*r {
		r = float32(w) / 2
	}
	if float32(h) < 2*r {
		r = float32(h) / 2
	}
	// 邊緣線（四邊）
	vector.StrokeLine(dst, float32(x)+r, float32(y), float32(x)+float32(w)-r, float32(y), width, col, true)
	vector.StrokeLine(dst, float32(x)+float32(w), float32(y)+r, float32(x)+float32(w), float32(y)+float32(h)-r, width, col, true)
	vector.StrokeLine(dst, float32(x)+r, float32(y)+float32(h), float32(x)+float32(w)-r, float32(y)+float32(h), width, col, true)
	vector.StrokeLine(dst, float32(x), float32(y)+r, float32(x), float32(y)+float32(h)-r, width, col, true)
	// 四角圓弧（以 StrokeCircle 畫整圓，視覺上可接受）
	vector.StrokeCircle(dst, float32(x)+r, float32(y)+r, r, width, col, true)
	vector.StrokeCircle(dst, float32(x)+float32(w)-r, float32(y)+r, r, width, col, true)
	vector.StrokeCircle(dst, float32(x)+r, float32(y)+float32(h)-r, r, width, col, true)
	vector.StrokeCircle(dst, float32(x)+float32(w)-r, float32(y)+float32(h)-r, r, width, col, true)
}

func canAfford(vm VM) bool {
	return vm.Research >= vm.NextUpgradeCost
}
