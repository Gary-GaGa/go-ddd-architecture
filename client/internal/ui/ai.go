package ui

import (
	"image/color"
	"time"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
	"github.com/hajimehoshi/ebiten/v2/vector"
	"golang.org/x/image/font"
)

type aiPalette struct {
	Node       color.RGBA
	NodeBorder color.RGBA
	Glow1      color.RGBA
	Glow2      color.RGBA
	Line       color.RGBA
	Spark      color.RGBA
	BgGlow     color.RGBA
	LabelBg    color.RGBA
}

// 更高解析的像素柴犬網格尺寸（可依需求調整 48/64）
const shibaGrid = 48

func paletteForLanguage(lang string) aiPalette {
	switch lang {
	case "py", "python":
		return aiPalette{Node: color.RGBA{255, 215, 80, 255}, NodeBorder: color.RGBA{170, 120, 0, 255}, Glow1: color.RGBA{255, 225, 120, 60}, Glow2: color.RGBA{255, 235, 160, 100}, Line: color.RGBA{100, 140, 220, 200}, Spark: color.RGBA{160, 200, 255, 200}, BgGlow: color.RGBA{60, 90, 150, 36}, LabelBg: color.RGBA{20, 40, 80, 200}}
	case "js", "javascript":
		return aiPalette{Node: color.RGBA{240, 255, 120, 255}, NodeBorder: color.RGBA{110, 130, 0, 255}, Glow1: color.RGBA{220, 255, 120, 60}, Glow2: color.RGBA{200, 255, 120, 100}, Line: color.RGBA{120, 200, 120, 200}, Spark: color.RGBA{180, 255, 160, 200}, BgGlow: color.RGBA{40, 90, 40, 36}, LabelBg: color.RGBA{20, 60, 20, 200}}
	default:
		return aiPalette{Node: color.RGBA{80, 180, 255, 255}, NodeBorder: color.RGBA{30, 90, 160, 255}, Glow1: color.RGBA{90, 200, 255, 50}, Glow2: color.RGBA{90, 200, 255, 90}, Line: color.RGBA{150, 150, 200, 200}, Spark: color.RGBA{200, 220, 255, 200}, BgGlow: color.RGBA{30, 70, 120, 36}, LabelBg: color.RGBA{10, 40, 80, 200}}
	}
}

// DrawEvolvingAI：柴犬視覺 + 小動畫
func DrawEvolvingAI(screen *ebiten.Image, face font.Face, x, y, w, h int, _ int64, lang string, _ int) {
	if w <= 0 || h <= 0 {
		return
	}
	pal := paletteForLanguage(lang)
	cx := float32(x + w/2)
	cy := float32(y + h/2)
	vector.DrawFilledCircle(screen, cx, cy, float32(min(w, h))/1.8, pal.BgGlow, true)

	// Pixel-art: draw to a small canvas then scale up with nearest filter
	// Choose a base logical resolution relative to area; snap to 24x24 grid multiples
	base := min(w, h)
	grid := shibaGrid
	k := base / grid
	if k < 4 {
		k = 4
	}
	if k > 8 {
		k = 8
	}
	pw := grid * k
	ph := pw
	if pw <= 0 || ph <= 0 {
		return
	}
	buf := ebiten.NewImage(pw, ph)
	// clear
	ebitenutil.DrawRect(buf, 0, 0, float64(pw), float64(ph), color.RGBA{0, 0, 0, 0})
	// draw shiba pixel-art into small buffer
	drawShiba(buf, face, 0, 0, pw, ph, pal, lang)
	// scale up to target rect with nearest-neighbor
	var op ebiten.DrawImageOptions
	sx := float64(w) / float64(pw)
	sy := float64(h) / float64(ph)
	op.GeoM.Scale(sx, sy)
	op.GeoM.Translate(float64(x), float64(y))
	op.Filter = ebiten.FilterNearest
	screen.DrawImage(buf, &op)
	// Draw language tag in screen space for crisp text
	if face != nil {
		tag := lang
		if tag == "" {
			tag = "go"
		}
		padX, padY := 6, 4
		textW := textWidth(face, tag)
		textH := 14
		tw := textW + padX*2
		th := textH + padY*2
		tagX := x + (w-tw)/2
		tagY := y + h - th - 8
		drawRoundedFilledRect(screen, tagX, tagY, tw, th, 6, Theme.CardBorder)
		drawRoundedFilledRect(screen, tagX+1, tagY+1, tw-2, th-2, 5, pal.LabelBg)
		drawText(screen, face, tag, tagX+padX, tagY+padY+12, color.RGBA{0, 0, 0, 255})
	}
}

// drawShiba 畫柴犬並掛上語言名牌；含呼吸、眨眼、尾巴擺動動畫。
func drawShiba(dst *ebiten.Image, _ font.Face, _ int, _ int, w, h int, pal aiPalette, _ string) {
	// 高解析像素網格
	gw, gh := shibaGrid, shibaGrid
	cellW := float64(w) / float64(gw)
	cellH := float64(h) / float64(gh)
	drawCell := func(cx, cy int, col color.RGBA) {
		if cx < 0 || cy < 0 || cx >= gw || cy >= gh {
			return
		}
		ebitenutil.DrawRect(dst, float64(cx)*cellW, float64(cy)*cellH, cellW, cellH, col)
	}

	// 構建符號網格
	grid := make([][]byte, gh)
	for y := 0; y < gh; y++ {
		grid[y] = make([]byte, gw)
		for x := 0; x < gw; x++ {
			grid[y][x] = '.'
		}
	}
	set := func(x, y int, ch byte) {
		if x >= 0 && y >= 0 && x < gw && y < gh {
			grid[y][x] = ch
		}
	}
	// 幾何填充
	fillCircle := func(cx, cy int, r float64, ch byte) {
		r2 := r * r
		for y := int(float64(cy) - r - 1); y <= int(float64(cy)+r+1); y++ {
			for x := int(float64(cx) - r - 1); x <= int(float64(cx)+r+1); x++ {
				dx, dy := float64(x-cx), float64(y-cy)
				if dx*dx+dy*dy <= r2 {
					set(x, y, ch)
				}
			}
		}
	}
	fillEllipse := func(cx, cy int, rx, ry float64, ch byte) {
		rx2, ry2 := rx*rx, ry*ry
		for y := int(float64(cy) - ry - 1); y <= int(float64(cy)+ry+1); y++ {
			for x := int(float64(cx) - rx - 1); x <= int(float64(cx)+rx+1); x++ {
				dx, dy := float64(x-cx), float64(y-cy)
				if (dx*dx)/rx2+(dy*dy)/ry2 <= 1 {
					set(x, y, ch)
				}
			}
		}
	}
	fillRect := func(x0, y0, x1, y1 int, ch byte) {
		if x0 > x1 {
			x0, x1 = x1, x0
		}
		if y0 > y1 {
			y0, y1 = y1, y0
		}
		for y := y0; y <= y1; y++ {
			for x := x0; x <= x1; x++ {
				set(x, y, ch)
			}
		}
	}
	imin := func(a, b int) int {
		if a < b {
			return a
		}
		return b
	}
	imax := func(a, b int) int {
		if a > b {
			return a
		}
		return b
	}
	fillTriangle := func(x0, y0, x1, y1, x2, y2 int, ch byte) {
		minX, maxX := imin(x0, imin(x1, x2)), imax(x0, imax(x1, x2))
		minY, maxY := imin(y0, imin(y1, y2)), imax(y0, imax(y1, y2))
		area := func(ax, ay, bx, by, cx, cy int) int { return (bx-ax)*(cy-ay) - (by-ay)*(cx-ax) }
		A := area(x0, y0, x1, y1, x2, y2)
		for y := minY; y <= maxY; y++ {
			for x := minX; x <= maxX; x++ {
				w0 := area(x1, y1, x2, y2, x, y)
				w1 := area(x2, y2, x0, y0, x, y)
				w2 := area(x0, y0, x1, y1, x, y)
				if (w0 >= 0 && w1 >= 0 && w2 >= 0 && A > 0) || (w0 <= 0 && w1 <= 0 && w2 <= 0 && A < 0) {
					set(x, y, ch)
				}
			}
		}
	}

	// 比例與定位
	cx, cy := gw/2, gh/2+gh/12
	R := float64(gw) * 0.34 // 頭半徑

	// 外框與毛色（保留 1px 輪廓）
	fillCircle(cx, cy, R, 'o')
	fillCircle(cx, cy, R-2, 'f')

	// 耳朵（外框 + 內填）
	earW := int(float64(gw) * 0.14)
	fillTriangle(cx-int(R*0.65), cy-int(R*0.78), cx-int(R*0.65)-earW, cy-int(R*0.35), cx-int(R*0.35), cy-int(R*0.35), 'o')
	fillTriangle(cx+int(R*0.65), cy-int(R*0.78), cx+int(R*0.65)+earW, cy-int(R*0.35), cx+int(R*0.35), cy-int(R*0.35), 'o')
	fillTriangle(cx-int(R*0.65), cy-int(R*0.78), cx-int(R*0.65)-earW+1, cy-int(R*0.35)+1, cx-int(R*0.35)+1, cy-int(R*0.35)+1, 'f')
	fillTriangle(cx+int(R*0.65), cy-int(R*0.78), cx+int(R*0.65)+earW-1, cy-int(R*0.35)+1, cx+int(R*0.35)-1, cy-int(R*0.35)+1, 'f')

	// 下側陰影
	fillEllipse(cx, cy+int(R*0.25), R*0.9, R*0.55, 'F')

	// 奶油白口鼻區
	fillEllipse(cx, cy+int(R*0.00), R*0.88, R*0.62, 'c')
	fillEllipse(cx, cy+int(R*0.10), R*0.74, R*0.50, 'c')

	// 眼睛與高光
	eyeDx := int(R * 0.46)
	eyeDy := int(-R * 0.04)
	eyeR := R * 0.085
	fillCircle(cx-eyeDx, cy+eyeDy, eyeR, 'b')
	fillCircle(cx+eyeDx, cy+eyeDy, eyeR, 'b')
	fillCircle(cx-eyeDx-1, cy+eyeDy-1, eyeR*0.4, 'h')
	fillCircle(cx+eyeDx-1, cy+eyeDy-1, eyeR*0.4, 'h')

	// 鼻頭與舌
	fillCircle(cx, cy+int(R*0.28), eyeR*0.7, 'n')
	fillRect(cx-int(eyeR*0.5), cy+int(R*0.32), cx+int(eyeR*0.5), cy+int(eyeR*0.35), 't')

	// 腮紅
	blushR := R * 0.12
	fillCircle(cx-int(R*0.55), cy+int(R*0.22), blushR, 'p')
	fillCircle(cx+int(R*0.55), cy+int(R*0.22), blushR, 'p')

	// 項圈（上亮下暗）
	colY := cy + int(R*0.85)
	fillRect(cx-int(R*1.0), colY, cx+int(R*1.0), colY+1, 'l')
	fillRect(cx-int(R*1.0), colY+2, cx+int(R*1.0), colY+3, 'd')

	// 眨眼
	now := time.Now()
	blinkClosed := (now.UnixMilli() % 4000) < 120

	// 顏色映射
	furCol := color.RGBA{214, 140, 60, 255}
	furDark := color.RGBA{174, 110, 45, 255}
	cream := color.RGBA{250, 238, 215, 255}
	black := color.RGBA{30, 30, 30, 255}
	outline := color.RGBA{80, 50, 30, 255}
	nose := black
	tongue := color.RGBA{220, 90, 90, 255}
	white := color.RGBA{255, 255, 255, 255}
	blush := color.RGBA{255, 150, 160, 255}

	// 繪製
	for y := 0; y < gh; y++ {
		for x := 0; x < gw; x++ {
			switch grid[y][x] {
			case 'o':
				drawCell(x, y, outline)
			case 'f':
				drawCell(x, y, furCol)
			case 'F':
				drawCell(x, y, furDark)
			case 'c':
				drawCell(x, y, cream)
			case 'b':
				if blinkClosed {
					drawCell(x, y, cream)
				} else {
					drawCell(x, y, black)
				}
			case 'h':
				drawCell(x, y, white)
			case 'n':
				drawCell(x, y, nose)
			case 't':
				drawCell(x, y, tongue)
			case 'p':
				drawCell(x, y, blush)
			case 'l':
				drawCell(x, y, pal.Node)
			case 'd':
				drawCell(x, y, pal.NodeBorder)
			}
		}
	}
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

// max 已在 hud.go 定義，這裡不重複定義。
