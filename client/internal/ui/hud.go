package ui

import (
	"fmt"
	"image/color"
	"math"
	"sort"
	"time"

	gametask "go-ddd-architecture/client/internal/game/task"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/text/v2"
	"github.com/hajimehoshi/ebiten/v2/vector"
	"golang.org/x/image/font"
)

// 任務區塊的場景渲染暫時停用（移除 mini game 與圓形倒數）

// 用於讓 App.Update() 取得本幀 Store Buy 按鈕的點擊區域
type uiRect struct{ X, Y, W, H int }

var (
	lastServerBuyRect uiRect
	lastGPUBuyRect    uiRect
)

// 防止單檔靜態檢查誤判未使用（跨檔案會使用）
func init() {
	_ = lastServerBuyRect
	_ = lastGPUBuyRect
}

// DrawHUD draws the main HUD: resources, level, current task timer, etc.
func DrawHUD(screen *ebiten.Image, face font.Face, vm VM, errMsg string, inFlight bool) {
	// background
	screen.Fill(Theme.Bg)

	b := screen.Bounds()
	sw, sh := b.Dx(), b.Dy()
	pad := Theme.Pad8
	// 內距至少 10px（P1 要求）
	innerPad := max(pad, 10)

	// layout widths（響應式調整，避免右側被切到）
	leftW := 280
	rightW := 260
	minLeft, minCenter, minRight := 220, 160, 200
	// 先以預設寬度計算 center；預留左右外側各 pad*2 與兩欄間隔 pad*2，共 pad*8
	centerW := sw - leftW - rightW - pad*8
	if centerW < minCenter {
		// 需要壓縮左右欄寬度以讓中欄至少保有最小寬度
		deficit := minCenter - centerW
		// 優先縮右欄
		shrinkR := min(deficit, rightW-minRight)
		rightW -= shrinkR
		deficit -= shrinkR
		// 再縮左欄
		if deficit > 0 {
			shrinkL := min(deficit, leftW-minLeft)
			leftW -= shrinkL
			deficit -= shrinkL
		}
		// 重新計算 centerW；若仍小於 minCenter，維持實際剩餘寬度以避免溢出
		centerW = sw - leftW - rightW - pad*8
	}

	// Left: Status card with light blue outline
	leftX, leftY := pad*2, pad*2
	leftH := 220
	drawCard(screen, leftX, leftY, leftW, leftH)
	// 額外淡藍描邊
	drawRoundedRectOutline(screen, leftX, leftY, leftW, leftH, Theme.Radius8, Theme.OutlineBlue, 1)
	tx, ty := leftX+innerPad, leftY+innerPad+14
	drawText(screen, face, "Status", tx, ty, Theme.TextMain)
	// Models badge: count languages as owned models
	models := len(vm.Languages)
	if models > 0 {
		mtxt := fmt.Sprintf("Models %d", models)
		mw := 6*2 + textWidth(face, mtxt)
		drawBadge(screen, leftX+leftW-innerPad-mw, ty-12, mtxt, face, true)
	}
	ty += 8
	vector.StrokeLine(screen, float32(tx), float32(ty), float32(leftX+leftW-innerPad), float32(ty), 1, Theme.CardBorder, true)
	ty += 12
	// Knowledge 行（含小幅跳動與飄字）
	drawText(screen, face, "Knowledge: ", tx, ty, Theme.TextMain)
	nK := fmt.Sprintf("%d", vm.Knowledge)
	kx := tx + textWidth(face, "Knowledge: ")
	ky := ty + bounceYOffset(vm.KBounceStart, vm.KBounceUntil, 3)
	// 放大強調（藍色）
	drawTextScaled(screen, face, nK, kx, ky, color.RGBA{0x58, 0xB8, 0xFF, 0xFF}, 1.25)
	// +K 飄字
	if vm.KFloatText != "" && !vm.KFloatStart.IsZero() && time.Now().Before(vm.KFloatUntil) {
		t := timeProgress(vm.KFloatStart, vm.KFloatUntil)
		fx := kx + textWidth(face, nK) + 10
		fy := ky - 6 - int(12*t)
		alpha := uint8(255 * (1 - t))
		col := color.RGBA{R: 0x58, G: 0xB8, B: 0xFF, A: alpha}
		drawText(screen, face, vm.KFloatText, fx, fy, col)
	}
	ty += 18
	// Research 行（含小幅跳動與飄字）
	drawText(screen, face, "Research:  ", tx, ty, Theme.TextMain)
	nR := fmt.Sprintf("%d", vm.Research)
	rx := tx + textWidth(face, "Research:  ")
	ry := ty + bounceYOffset(vm.RBounceStart, vm.RBounceUntil, 3)
	// 放大強調（綠色）
	drawTextScaled(screen, face, nR, rx, ry, color.RGBA{0x6B, 0xE5, 0x9C, 0xFF}, 1.25)
	// +R 飄字
	if vm.RFloatText != "" && !vm.RFloatStart.IsZero() && time.Now().Before(vm.RFloatUntil) {
		t := timeProgress(vm.RFloatStart, vm.RFloatUntil)
		fx := rx + textWidth(face, nR) + 10
		fy := ry - 6 - int(12*t)
		alpha := uint8(255 * (1 - t))
		col := color.RGBA{R: 0x6B, G: 0xE5, B: 0x9C, A: alpha}
		drawText(screen, face, vm.RFloatText, fx, fy, col)
	}
	ty += 18
	drawText(screen, face, "Level: ", tx, ty, Theme.TextMain)
	nLv := fmt.Sprintf("%d", vm.Level)
	// 放大強調（橘色）
	drawTextScaled(screen, face, nLv, tx+textWidth(face, "Level: "), ty, color.RGBA{0xFF, 0xB3, 0x6B, 0xFF}, 1.25)
	costStr := fmt.Sprintf("Cost R %d", vm.NextUpgradeCost)
	badgeW := 6*2 + textWidth(face, costStr)
	drawBadge(screen, leftX+leftW-pad-badgeW, ty-12, costStr, face, canAfford(vm))
	ty += 22
	drawText(screen, face, fmt.Sprintf("Rates K/R per min: %d / %d", vm.KnowledgePerMin, vm.ResearchPerMin), tx, ty, Theme.TextSub)
	ty += 18
	// Hardware bonus: GPUs increase R/min
	bonus := vm.GPUs * GPUBonusRPM
	drawText(screen, face, fmt.Sprintf("HW bonus R/min: +%d (GPUs %d x %d)", bonus, vm.GPUs, GPUBonusRPM), tx, ty, Theme.TextSub)
	ty += 18
	if vm.EstimatedSuccess > 0 {
		drawText(screen, face, fmt.Sprintf("Est. Success: %.0f%%", vm.EstimatedSuccess*100), tx, ty, Theme.TextSub)
		ty += 22
	}

	// Networking / Error in left card bottom area
	nx := leftX + innerPad
	ny := leftY + leftH - innerPad - 36
	if inFlight {
		// "Networking..." with bouncing dots
		base := "Networking"
		drawText(screen, face, base, nx, ny, Theme.Good)
		dotsX := nx + textWidth(face, base) + 8
		dotsY := ny - 6
		now := time.Now()
		period := float64(1200 * time.Millisecond)
		for i := 0; i < 3; i++ {
			t := float64(now.UnixNano()%int64(period)) / period
			phase := t + float64(i)/3 // 相位差
			yoff := -2.5 * math.Sin(2*math.Pi*phase)
			x := float32(dotsX + i*8)
			vector.DrawFilledCircle(screen, x, float32(dotsY)+float32(yoff), 2, Theme.Good, true)
		}
		ny += 18
	}
	if errMsg != "" {
		drawText(screen, face, "Error: "+errMsg, nx, ny, Theme.Error)
		ny += 18
	}

	// 取得滑鼠位置，供 hover 判斷
	mx, my := ebiten.CursorPosition()

	// Center: Unified Task block（取代原進度條與獨立戰鬥視覺）
	centerX := leftX + leftW + pad*2
	centerY := leftY
	// 以左側 Status+Store 高度為基準底線
	baselineBottom := leftY + leftH + pad + 160 // 160 為 Store 高度
	// 不預留 Hotkeys 橫條高度，Task 直接拉到底
	taskH := baselineBottom - centerY
	if taskH < 160 {
		taskH = 160
	}
	drawCard(screen, centerX, centerY, centerW, taskH)
	drawRoundedRectOutline(screen, centerX, centerY, centerW, taskH, Theme.Radius8, Theme.OutlineBlue, 1)
	// Blueprint background grid (subtle)
	drawBlueprintGrid(screen, centerX+1, centerY+1, centerW-2, taskH-2)
	ttx, tty := centerX+innerPad, centerY+innerPad+12
	if vm.CurrentTask != nil {
		title := "Task: " + vm.CurrentTask.Type
		if vm.CurrentTask.Language != "" {
			title += " [" + vm.CurrentTask.Language + "]"
		}
		drawText(screen, face, title, ttx, tty, Theme.TextMain)
		// Type tag at top-right of task card
		if vm.CurrentTask.Type != "" {
			tag := vm.CurrentTask.Type
			tw := 12 + textWidth(face, tag)
			drawBadge(screen, centerX+centerW-pad-tw, centerY+pad, tag, face, true)
		}
		tty += 16
		if vm.CurrentTask.Language != "" {
			startLabel := "Started: " + vm.CurrentTask.Language
			if vm.CurrentLanguage != "" && vm.CurrentLanguage != vm.CurrentTask.Language {
				startLabel += " (now: " + vm.CurrentLanguage + ")"
			}
			drawText(screen, face, startLabel, ttx, tty, Theme.Good)
			tty += 16
		} else {
			tty += 2
		}
		// P2: 顯示進度提示
		if vm.CurrentTask.BaseReward > 0 {
			hint := fmt.Sprintf("Task in progress: +%dK", vm.CurrentTask.BaseReward)
			drawText(screen, face, hint, ttx, tty, Theme.TextSub)
			tty += 16
		}
		// P1: 小遊戲進度（鑰匙 -> 目標方塊）
		remain := vm.CurrentTask.RemainingSeconds
		total := vm.CurrentTask.DurationSeconds
		var pct float32
		if total > 0 {
			elapsed := total - remain
			if elapsed < 0 {
				elapsed = 0
			}
			pct = float32(elapsed) / float32(total)
			if pct < 0 {
				pct = 0
			} else if pct > 1 {
				pct = 1
			}
		}
		// 區域（渲染 Practice MVP，不使用圓形倒數）
		areaX := centerX + innerPad
		areaY := centerY + innerPad + 64
		areaW := centerW - innerPad*2
		areaH := taskH - (areaY - centerY) - pad
		// 建立渲染器（暫時每幀建構輕量物件，之後可升級為單例）
		var renderer gametask.Renderer
		// 先看本地預覽覆寫（若有）
		rtype := vm.CurrentTask.Type
		if vm.TaskPreview != "" {
			rtype = vm.TaskPreview
		}
		switch rtype {
		case "deploy", "Deploy", "DEPLOY":
			renderer = gametask.NewDeployRenderer()
		case "research", "Research", "RESEARCH":
			renderer = gametask.NewResearchRenderer()
		default:
			renderer = gametask.NewPracticeRenderer()
		}
		renderer.Sync(vm.CurrentTask.DurationSeconds, vm.CurrentTask.RemainingSeconds, vm.CurrentTask.BaseReward)
		renderer.Update(16 * time.Millisecond)
		// Theme 映射
		theme := gametask.Theme{
			Grid:   Theme.CardBorder,
			Key:    Theme.TextMain,
			Goal:   Theme.TextSub,
			Bonus:  Theme.Accent,
			Text:   Theme.TextSub,
			Accent: Theme.Accent,
			Good:   Theme.Good,
			Warn:   Theme.Warn,
		}
		renderer.Draw(screen, areaX, areaY, areaW, areaH, theme)
	} else {
		drawText(screen, face, "No active task", ttx, tty, Theme.TextSub)
		// Offer Practice button (mouse hover supported)
		btnW := 14 + textWidth(face, "Practice")
		btnH := 24
		bx := ttx
		by := tty + 10
		hovered := pointInRect(mx, my, bx, by-16, btnW, btnH)
		drawButton(screen, bx, by-16, btnW, btnH, "Practice", face, true, hovered)
		// 若有本地預覽，仍渲染預覽內容（deploy/research skeleton）
		if vm.TaskPreview != "" {
			areaX := centerX + innerPad
			areaY := centerY + innerPad + 64
			areaW := centerW - innerPad*2
			areaH := taskH - (areaY - centerY) - pad
			var renderer gametask.Renderer
			switch vm.TaskPreview {
			case "deploy":
				renderer = gametask.NewDeployRenderer()
			case "research":
				renderer = gametask.NewResearchRenderer()
			default:
				renderer = gametask.NewPracticeRenderer()
			}
			renderer.Sync(20, 10, 5)
			renderer.Update(16 * time.Millisecond)
			theme := gametask.Theme{Grid: Theme.CardBorder, Key: Theme.TextMain, Goal: Theme.TextSub, Bonus: Theme.Accent, Text: Theme.TextSub, Accent: Theme.Accent, Good: Theme.Good, Warn: Theme.Warn}
			renderer.Draw(screen, areaX, areaY, areaW, areaH, theme)
		}
	}

	// 已移除獨立戰鬥區塊：進度動畫已整合於 Task 區塊

	// 移除 Task 底下 Hotkeys 橫條（統一在畫面最底顯示提示）

	// Right: Languages panel（延伸到底）
	rightX := centerX + centerW + pad*2
	rightY := leftY
	// 右欄保證不超出視窗右側安全邊距 pad*2
	if rxMax := sw - pad*2 - rightX; rxMax < rightW {
		if rxMax < 0 {
			rightW = 0
		} else {
			rightW = rxMax
		}
	}
	baselineBottom = leftY + leftH + pad + 160 // 與左側一致
	// 增加列高，分成三行（標題/數值/進度條）以避免擁擠
	langRowH := 54
	ribbonH := 28
	langH := baselineBottom - rightY
	drawCard(screen, rightX, rightY, rightW, langH)
	lx, ly := rightX+pad, rightY+pad+12
	drawText(screen, face, "Languages", lx, ly, Theme.TextMain)
	ly += 16
	// 建立語言集（確保 go/py/js 存在）
	langs := map[string]LangHUD{}
	for k, v := range vm.Languages {
		langs[k] = v
	}
	for _, code := range []string{"go", "py", "js"} {
		if _, ok := langs[code]; !ok {
			langs[code] = LangHUD{Code: code}
		}
	}
	// 依模式排序
	keys := make([]string, 0, len(langs))
	for k := range langs {
		keys = append(keys, k)
	}
	mode := vm.LangSort
	if mode == "" {
		mode = "lv"
	}
	sort.Slice(keys, func(i, j int) bool {
		si := langs[keys[i]]
		sj := langs[keys[j]]
		switch mode {
		case "k":
			if si.Knowledge != sj.Knowledge {
				return si.Knowledge > sj.Knowledge
			}
			if si.Level != sj.Level {
				return si.Level > sj.Level
			}
		case "recent":
			ti := int64(0)
			tj := int64(0)
			if vm.LangRecent != nil {
				if v, ok := vm.LangRecent[keys[i]]; ok {
					ti = v
				}
				if v, ok := vm.LangRecent[keys[j]]; ok {
					tj = v
				}
			}
			if ti != tj {
				return ti > tj
			}
			if si.Level != sj.Level {
				return si.Level > sj.Level
			}
			if si.Knowledge != sj.Knowledge {
				return si.Knowledge > sj.Knowledge
			}
		default: // lv
			if si.Level != sj.Level {
				return si.Level > sj.Level
			}
			if si.Knowledge != sj.Knowledge {
				return si.Knowledge > sj.Knowledge
			}
		}
		return keys[i] < keys[j]
	})
	// track best next unlock
	bestCode := ""
	var bestRem int64 = -1
	row := func(code string, s LangHUD) {
		// mini-card background per language row
		rx0 := rightX + 6
		ry0 := ly - 14
		rw0 := rightW - 12
		rh0 := langRowH - 8
		drawRoundedFilledRect(screen, rx0, ry0, rw0, rh0, 6, Theme.CardBg)
		drawRoundedRectOutline(screen, rx0, ry0, rw0, rh0, 6, Theme.CardBorder, 1)
		// 第一行：語言代碼與 READY
		col := Theme.TextSub
		if code == vm.CurrentLanguage {
			col = Theme.Good
		}
		titleY := ly - 2
		drawText(screen, face, code, lx, titleY, col)
		// READY 標示（用 badge 顯示）
		nextCost := nextUpgradeCostForLevel(s.Level)
		ready := s.Research >= nextCost && nextCost > 0
		if ready {
			// 將 READY 放在語言代碼之後，避免壓到右側 K/R/Lv
			codeEndX := lx + textWidth(face, code)
			badgeX := codeEndX + 8
			badgeY := titleY - 10
			drawBadge(screen, badgeX, badgeY, "READY", face, true)
		}
		// 第二行：K / R（靠左）與 Lv（靠右）
		statsY := ly + 14
		bx := lx
		// K:
		drawText(screen, face, "K:", bx, statsY, col)
		bx += textWidth(face, "K:")
		drawText(screen, face, fmt.Sprintf("%d", s.Knowledge), bx, statsY, color.RGBA{0x58, 0xB8, 0xFF, 0xFF})
		bx += textWidth(face, fmt.Sprintf("%d", s.Knowledge))
		// spacing + R:
		drawText(screen, face, "   R:", bx, statsY, col)
		bx += textWidth(face, "   R:")
		drawText(screen, face, fmt.Sprintf("%d", s.Research), bx, statsY, color.RGBA{0x6B, 0xE5, 0x9C, 0xFF})
		// Lv 右對齊
		lvLabel := fmt.Sprintf("Lv:%d", s.Level)
		lvW := textWidth(face, lvLabel)
		lvX := rightX + rightW - pad - lvW
		drawText(screen, face, lvLabel, lvX, statsY, col)
		// 取消右上小錠片，避免壓到右側資訊
		// upgrade glow: outline glow around the mini-card for 1s
		if vm.Pulses != nil {
			if until, ok := vm.Pulses[code]; ok {
				left := time.Until(until)
				if left > 0 {
					pulseWindow := 1000 * time.Millisecond
					if left > pulseWindow {
						left = pulseWindow
					}
					// t in [0,1]
					t := 1 - float32(left)/float32(pulseWindow)
					if t < 0 {
						t = 0
					} else if t > 1 {
						t = 1
					}
					a := uint8(140 + 100*(1-t))
					glow := color.RGBA{R: Theme.Accent.R, G: Theme.Accent.G, B: Theme.Accent.B, A: a}
					drawRoundedRectOutline(screen, rx0-1, ry0-1, rw0+2, rh0+2, 7, glow, 2)
				}
			}
		}
		// progress bar line (Research toward next upgrade)
		if nextCost <= 0 {
			nextCost = 1
		}
		pct := float32(0)
		if s.Research > 0 {
			pct = float32(s.Research) / float32(nextCost)
			if pct > 1 {
				pct = 1
			}
		}
		barX := rx0 + 8
		barY := ly + 28
		barW := rightW - pad*2
		barH := 8
		// background track with subtle alpha
		track := color.RGBA{R: Theme.CardBorder.R, G: Theme.CardBorder.G, B: Theme.CardBorder.B, A: 140}
		fill := Theme.Good
		if ready {
			fill = Theme.Accent
		}
		if barW > 0 {
			drawProgressBar(screen, barX, barY, barW, barH, pct, track, fill)
		}
		// 右對齊提示 "R x / y"
		hint := fmt.Sprintf("R %d / %d", s.Research, nextCost)
		hx := barX + barW - textWidth(face, hint)
		hy := barY + barH - 2
		drawText(screen, face, hint, hx, hy, Theme.TextSub)
		// record next unlock candidate
		if !ready {
			rem := nextCost - s.Research
			if rem < 0 {
				rem = 0
			}
			if bestRem < 0 || rem < bestRem {
				bestRem = rem
				bestCode = code
			}
		}
		ly += langRowH
	}
	// 依排序 keys 逐行畫出
	for _, k := range keys {
		row(k, langs[k])
	}
	// Next Unlock ribbon at bottom（改為復古黑底綠字螢幕風格）
	ribX := rightX + 1
	ribY := rightY + langH - ribbonH - 1
	ribW := rightW - 2
	ribH := ribbonH - 6
	crtBg := color.RGBA{0x05, 0x08, 0x0A, 0xFF} // 近黑底
	neon := color.RGBA{0x4D, 0xF7, 0x7A, 0xCC}  // 螢光綠邊
	txt := color.RGBA{0x4D, 0xF7, 0x7A, 0xFF}   // 螢光綠字
	drawRoundedFilledRect(screen, ribX+4, ribY+2, ribW-8, ribH, 6, crtBg)
	drawRoundedRectOutline(screen, ribX+4, ribY+2, ribW-8, ribH, 6, neon, 1.2)
	// 螢幕掃描線
	for hy := ribY + 3; hy < ribY+2+ribH-2; hy += 3 {
		c := color.RGBA{txt.R, txt.G, txt.B, 35}
		vector.StrokeLine(screen, float32(ribX+6), float32(hy), float32(ribX+ribW-6), float32(hy), 1, c, true)
	}
	info := ""
	if bestCode != "" && bestRem >= 0 {
		info = fmt.Sprintf("> Next: %s +Lv (need R %d)", bestCode, bestRem)
	} else {
		info = "> Next: Ready to upgrade!"
	}
	drawText(screen, face, info, ribX+8, ribY+2+ribH/2+6, txt)

	// Bottom-left: Store panel (Servers / GPUs)
	// layout: under Status card at left
	storeX := leftX
	storeY := leftY + leftH + pad
	storeW := leftW
	storeH := 160
	drawCard(screen, storeX, storeY, storeW, storeH)
	drawRoundedRectOutline(screen, storeX, storeY, storeW, storeH, Theme.Radius8, Theme.OutlineBlue, 1)
	stx, sty := storeX+innerPad, storeY+innerPad+12
	drawText(screen, face, "Store", stx, sty, Theme.TextMain)
	sty += 12
	// mini cards: Server & GPU
	miniW := (storeW - pad*3) / 2
	miniH := 96
	inner := 10
	// Server card
	srvX := storeX + innerPad
	srvY := sty
	drawRoundedFilledRect(screen, srvX, srvY, miniW, miniH, 8, Theme.CardBg)
	drawRoundedRectOutline(screen, srvX, srvY, miniW, miniH, 8, Theme.CardBorder, 1)
	// line 1: title
	drawText(screen, face, "Server", srvX+inner, srvY+inner+12, Theme.TextMain)
	// line 2: owned
	drawText(screen, face, fmt.Sprintf("Owned: %d", vm.Servers), srvX+inner, srvY+inner+12+16, Theme.TextSub)
	// line 3: price
	costServer := fmt.Sprintf("K %d", StoreCostServerK)
	drawText(screen, face, "Price:", srvX+inner, srvY+inner+12+16+16, Theme.TextSub)
	drawText(screen, face, costServer, srvX+inner+textWidth(face, "Price:")+6, srvY+inner+12+16+16, color.RGBA{0x58, 0xB8, 0xFF, 0xFF})
	// line 4: buy button at bottom with padding
	canS := vm.CanAffordServer()
	btnSW := 14 + textWidth(face, "Buy")
	btnSH := 22
	btnSX := srvX + inner
	btnSY := srvY + miniH - inner - btnSH
	hoveredS := pointInRect(mx, my, btnSX, btnSY, btnSW, btnSH)
	drawButton(screen, btnSX, btnSY, btnSW, btnSH, "Buy", face, canS, hoveredS)
	// 記錄本幀 Server Buy 點擊矩形
	lastServerBuyRect = uiRect{X: btnSX, Y: btnSY, W: btnSW, H: btnSH}
	// GPU card
	gpuX := srvX + miniW + pad
	gpuY := sty
	drawRoundedFilledRect(screen, gpuX, gpuY, miniW, miniH, 8, Theme.CardBg)
	drawRoundedRectOutline(screen, gpuX, gpuY, miniW, miniH, 8, Theme.CardBorder, 1)
	// line 1: title
	drawText(screen, face, "GPU", gpuX+inner, gpuY+inner+12, Theme.TextMain)
	// line 2: owned
	drawText(screen, face, fmt.Sprintf("Owned: %d", vm.GPUs), gpuX+inner, gpuY+inner+12+16, Theme.TextSub)
	// line 3: price
	costGPU := fmt.Sprintf("K %d", StoreCostGPUK)
	drawText(screen, face, "Price:", gpuX+inner, gpuY+inner+12+16+16, Theme.TextSub)
	drawText(screen, face, costGPU, gpuX+inner+textWidth(face, "Price:")+6, gpuY+inner+12+16+16, color.RGBA{0x58, 0xB8, 0xFF, 0xFF})
	// line 4: buy button
	canG := vm.CanAffordGPU()
	btnGW := 14 + textWidth(face, "Buy")
	btnGH := 22
	btnGX := gpuX + inner
	btnGY := gpuY + miniH - inner - btnGH
	hoveredG := pointInRect(mx, my, btnGX, btnGY, btnGW, btnGH)
	drawButton(screen, btnGX, btnGY, btnGW, btnGH, "Buy", face, canG, hoveredG)
	// 記錄本幀 GPU Buy 點擊矩形
	lastGPUBuyRect = uiRect{X: btnGX, Y: btnGY, W: btnGW, H: btnGH}
	// slots bar below
	sty = sty + miniH + 8
	drawSlotsBar(screen, stx, sty, vm.Slots, vm.GPUs, vm)
	// no extra hint here; see bottom bar

	// 底部極簡 Hotkeys（僅必要鍵，並尊重 ShowHotkeys）
	if vm.ShowHotkeys {
		// 幾個層級，依螢幕寬度自適應
		full := "P Practice  T Targeted  D Deploy  R Research  U Upgrade  C Claim  1 Go  2 Py  3 JS"
		mid := "P T D R U C  |  1 Go 2 Py 3 JS"
		small := "P T D R U C | 1 2 3"
		// 選擇可容納的字串
		candidates := []string{full, mid, small}
		chosen := small
		for _, s := range candidates {
			if textWidth(face, s) <= sw-Theme.Pad8*2 {
				chosen = s
				break
			}
		}
		bw := textWidth(face, chosen)
		cx := (sw - bw) / 2
		drawText(screen, face, chosen, cx, sh-Theme.Pad8, Theme.TextSub)
	}
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
	// Pulses 短暫脈動效果：語言代碼 -> 截止時間（在此之前顯示一圈擴散環）
	Pulses map[string]time.Time
	// 戰鬥動畫：任務成功/失敗的短暫呈現
	BattleType  string
	BattleUntil time.Time
	BattleLang  string
	BattleStyle int
	// Store
	Servers int
	GPUs    int
	Slots   int
	// Animations (ephemeral, provided by App)
	KBounceStart    time.Time
	KBounceUntil    time.Time
	RBounceStart    time.Time
	RBounceUntil    time.Time
	KFloatText      string
	KFloatStart     time.Time
	KFloatUntil     time.Time
	RFloatText      string
	RFloatStart     time.Time
	RFloatUntil     time.Time
	GPUAnimStart    time.Time
	GPUAnimUntil    time.Time
	ServerAnimStart time.Time
	ServerAnimUntil time.Time
	// UI toggles
	ShowHotkeys bool
	// 語言排序與最近使用
	LangSort   string
	LangRecent map[string]int64 // 語言 -> 最近使用（unix 秒）
	// 本地任務視覺預覽（不影響後端）："deploy" | "research" | ""（空=依照後端）
	TaskPreview string
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
	drawRoundedFilledRect(dst, x, y, w, h, Theme.Radius8, Theme.CardBg)
}

func drawBadge(dst *ebiten.Image, x, y int, textStr string, face font.Face, ok bool) {
	padX, padY := 6, 4
	w := padX*2 + textWidth(face, textStr)
	h := padY*2 + 14
	// 背景色：可負擔用 Good，不可負擔用 Warn，並做小幅脈動（提示不足）
	var bg color.RGBA
	if ok {
		bg = Theme.Good
	} else {
		period := float64(1200 * time.Millisecond)
		t := float64(time.Now().UnixNano()%int64(period)) / period
		a := uint8(180 + 75*0.5*(1+math.Sin(2*math.Pi*t)))
		bg = color.RGBA{R: Theme.Warn.R, G: Theme.Warn.G, B: Theme.Warn.B, A: a}
	}
	drawRoundedFilledRect(dst, x, y, w, h, 6, Theme.CardBorder)
	drawRoundedFilledRect(dst, x+1, y+1, w-2, h-2, 5, bg)
	drawText(dst, face, textStr, x+padX, y+padY+12, color.RGBA{0, 0, 0, 255})
}

// --- small helpers for animations ---

func timeProgress(start, until time.Time) float32 {
	if start.IsZero() || until.IsZero() {
		return 0
	}
	now := time.Now()
	if now.After(until) {
		return 1
	}
	total := until.Sub(start).Seconds()
	if total <= 0 {
		return 1
	}
	t := now.Sub(start).Seconds() / total
	if t < 0 {
		t = 0
	}
	if t > 1 {
		t = 1
	}
	return float32(t)
}

// bounceYOffset returns a small negative y offset (upwards) with ease-out sinus curve.
func bounceYOffset(start, until time.Time, amp int) int {
	if start.IsZero() || until.IsZero() {
		return 0
	}
	t := timeProgress(start, until)
	// use a single up-down arc: y = -sin(t*pi) * amp
	off := -math.Sin(float64(t)*math.Pi) * float64(amp)
	return int(off)
}

// drawSlotsBar draws slot frames and filled GPUs; uses vm GPU/Server animations to render effects.
func drawSlotsBar(dst *ebiten.Image, x, y int, slots int, gpus int, vm VM) {
	if slots <= 0 {
		return
	}
	gap := 6
	w, h := 12, 10
	// pulse glow when server just added
	pulse := false
	if !vm.ServerAnimStart.IsZero() && time.Now().Before(vm.ServerAnimUntil) {
		pulse = true
	}
	for i := 0; i < slots; i++ {
		sx := x + i*(w+gap)
		sy := y
		border := Theme.CardBorder
		if pulse {
			// subtle glow
			t := timeProgress(vm.ServerAnimStart, vm.ServerAnimUntil)
			a := uint8(120 + 80*(1-t))
			border = color.RGBA{R: Theme.OutlineBlue.R, G: Theme.OutlineBlue.G, B: Theme.OutlineBlue.B, A: a}
		}
		drawRoundedRectOutline(dst, sx, sy, w, h, 2, border, 1)
		if i < gpus {
			drawRoundedFilledRect(dst, sx+1, sy+1, w-2, h-2, 1.5, Theme.Good)
		}
	}
	// GPU insert animation: slide a GPU into the last filled slot
	if gpus > 0 && !vm.GPUAnimStart.IsZero() && time.Now().Before(vm.GPUAnimUntil) {
		t := timeProgress(vm.GPUAnimStart, vm.GPUAnimUntil)
		idx := gpus - 1
		if idx >= slots {
			idx = slots - 1
		}
		sx := x + idx*(w+gap)
		sy := y
		// start above and slide down
		yy := float32(sy - 14 + int(14*t))
		drawRoundedFilledRect(dst, sx+1, int(yy)+1, w-2, h-2, 1.5, Theme.Good)
		// small glow at the end
		glowA := uint8(200 * (1 - t))
		glow := color.RGBA{R: Theme.Good.R, G: Theme.Good.G, B: Theme.Good.B, A: glowA}
		vector.StrokeRect(dst, float32(sx), float32(sy), float32(w), float32(h), 2, glow, true)
	}
}

// 移除圓角弧環與內插相關輔助（目前未使用）

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}

// --- rounded rect helpers ---

func drawRoundedFilledRect(dst *ebiten.Image, x, y, w, h int, _ float32, col color.Color) {
	// 改為純方角填滿，移除四角圓圈
	if w <= 0 || h <= 0 {
		return
	}
	vector.DrawFilledRect(dst, float32(x), float32(y), float32(w), float32(h), col, true)
}

func drawRoundedRectOutline(dst *ebiten.Image, x, y, w, h int, _ float32, col color.Color, width float32) {
	// 改為純方角外框，移除四角圓圈
	if w <= 0 || h <= 0 {
		return
	}
	vector.StrokeRect(dst, float32(x), float32(y), float32(w), float32(h), width, col, true)
}

func canAfford(vm VM) bool {
	return vm.Research >= vm.NextUpgradeCost
}

// drawText 是 text/v2 的薄封裝，提供舊介面形式的呼叫方式。
func drawText(dst *ebiten.Image, face font.Face, s string, x, y int, col color.RGBA) {
	xf := text.NewGoXFace(face)
	m := xf.Metrics()
	// 模擬 v1：以 (x,y) 為 baseline，將繪製區塊往上移動 HAscent
	var op text.DrawOptions
	op.GeoM.Translate(float64(x), float64(y)-m.HAscent)
	op.ColorScale.ScaleWithColor(col)
	text.Draw(dst, s, xf, &op)
}

// drawTextScaled draws text with a scale factor and correct baseline.
func drawTextScaled(dst *ebiten.Image, face font.Face, s string, x, y int, col color.RGBA, scale float64) {
	if scale <= 0 {
		scale = 1
	}
	xf := text.NewGoXFace(face)
	m := xf.Metrics()
	var op text.DrawOptions
	op.GeoM.Scale(scale, scale)
	// translate after scale: baseline at y - HAscent*scale
	op.GeoM.Translate(float64(x), float64(y)-m.HAscent*scale)
	op.ColorScale.ScaleWithColor(col)
	text.Draw(dst, s, xf, &op)
}

// textWidth 使用 text/v2 測量文字寬度（像素，向上取整）。
func textWidth(face font.Face, s string) int {
	if s == "" || face == nil {
		return 0
	}
	xf := text.NewGoXFace(face)
	w, _ := text.Measure(s, xf, 0)
	return int(math.Ceil(w))
}

// --- Store 常數（與後端同步）：成本（Knowledge）
const (
	StoreCostServerK = 150
	StoreCostGPUK    = 80
	// 顯卡對研究產率的固定加成（顯示用途；後端目前為 +1 R/min/顯卡）
	GPUBonusRPM = 1
)

// 推導可負擔（純前端估計，不作為實際判定）：
func (vm VM) CanAffordServer() bool { return vm.Knowledge >= StoreCostServerK }

// GPU 還需插槽可用
func (vm VM) CanAffordGPU() bool {
	if vm.Knowledge < StoreCostGPUK {
		return false
	}
	return vm.GPUs < vm.Slots
}

// drawBlueprintGrid renders a subtle moving grid background.
func drawBlueprintGrid(dst *ebiten.Image, x, y, w, h int) {
	if w <= 0 || h <= 0 {
		return
	}
	step := 12
	// slight horizontal animation based on time
	t := float32(time.Now().UnixNano()%1_000_000_000) / 1_000_000_000
	off := int(float32(step) * (t * 0.5)) // slow drift
	gridCol := color.RGBA{R: Theme.CardBorder.R, G: Theme.CardBorder.G, B: Theme.CardBorder.B, A: 60}
	// vertical lines
	for vx := x + off; vx < x+w; vx += step {
		vector.StrokeLine(dst, float32(vx), float32(y), float32(vx), float32(y+h), 1, gridCol, true)
	}
	// horizontal lines
	for hy := y + off; hy < y+h; hy += step {
		vector.StrokeLine(dst, float32(x), float32(hy), float32(x+w), float32(hy), 1, gridCol, true)
	}
}

// drawPillChip：已取消晶片樣式（保留註解位置給未來擴充）。

// nextUpgradeCostForLevel replicates the server's cost formula for display: 100 * 2^level.
func nextUpgradeCostForLevel(level int) int64 {
	cost := int64(100)
	for i := 0; i < level; i++ {
		cost *= 2
	}
	return cost
}

// drawProgressBar draws a simple progress bar with a track and a filled portion.
func drawProgressBar(dst *ebiten.Image, x, y, w, h int, pct float32, track color.RGBA, fill color.RGBA) {
	if w <= 0 || h <= 0 {
		return
	}
	if pct < 0 {
		pct = 0
	}
	if pct > 1 {
		pct = 1
	}
	// track
	drawRoundedFilledRect(dst, x, y, w, h, 4, track)
	// fill
	fw := int(float32(w) * pct)
	if fw > 0 {
		drawRoundedFilledRect(dst, x, y, fw, h, 4, fill)
	}
}

// --- button helpers ---

// pointInRect returns whether (mx,my) is inside rect x,y,w,h
func pointInRect(mx, my, x, y, w, h int) bool {
	return mx >= x && my >= y && mx < x+w && my < y+h
}

// drawButton draws a pill-like button with hover and disabled states.
// When disabled, it draws a lock icon and grays out the content.
func drawButton(dst *ebiten.Image, x, y, w, h int, label string, face font.Face, enabled bool, hovered bool) {
	if w <= 0 || h <= 0 {
		return
	}
	radius := float32(h) / 2
	// base colors
	var bg color.RGBA
	fg := Theme.TextMain
	var border color.RGBA
	if enabled {
		bg = color.RGBA{R: Theme.Accent.R, G: Theme.Accent.G, B: Theme.Accent.B, A: 40}
		border = color.RGBA{R: Theme.Accent.R, G: Theme.Accent.G, B: Theme.Accent.B, A: 150}
	} else {
		// 灰階
		g := color.RGBA{0x7A, 0x86, 0x99, 0xFF}
		bg = color.RGBA{R: g.R, G: g.G, B: g.B, A: 40}
		border = g
		fg = color.RGBA{0xBD, 0xC7, 0xD4, 0xFF}
	}
	// 按下效果：縮小 1px 邊距（僅視覺，簡單版）
	if enabled && hovered && ebiten.IsMouseButtonPressed(ebiten.MouseButtonLeft) {
		x += 1
		y += 1
		w -= 2
		h -= 2
	}
	// draw body
	drawRoundedFilledRect(dst, x, y, w, h, radius, bg)
	drawRoundedRectOutline(dst, x, y, w, h, radius, border, 1)
	// glow on hover when enabled
	if enabled && hovered {
		// 改為矩形外框高光，避免角落圓圈效果
		glow := color.RGBA{R: Theme.OutlineBlue.R, G: Theme.OutlineBlue.G, B: Theme.OutlineBlue.B, A: 200}
		vector.StrokeRect(dst, float32(x), float32(y), float32(w), float32(h), 2, glow, true)
	}
	// label
	tx := x + (w-textWidth(face, label))/2
	ty := y + h/2 + 6
	drawText(dst, face, label, tx, ty, fg)
	// disabled lock icon
	if !enabled {
		// small lock at left side of the label
		lx := x + 6
		ly := y + h/2 - 6
		lockCol := border
		// shackle (arc-like using circles/lines)
		vector.StrokeCircle(dst, float32(lx+6), float32(ly+6), 6, 2, lockCol, true)
		vector.DrawFilledRect(dst, float32(lx), float32(ly+6), 12, 10, lockCol, true)
		keyhole := color.RGBA{0, 0, 0, 180}
		vector.DrawFilledCircle(dst, float32(lx+6), float32(ly+11), 1.5, keyhole, true)
	}
}

// （保留位置給未來 mini-game 形狀輔助）
