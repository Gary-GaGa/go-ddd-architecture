package task

import (
	"fmt"
	"math"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/text/v2"
	"github.com/hajimehoshi/ebiten/v2/vector"
	"golang.org/x/image/font"
)

// DrawHUDWidgets 繪製任務 HUD（倒數字與圓形進度圈）。
func (s *Scene) DrawHUDWidgets(dst *ebiten.Image, face font.Face, originX, originY int, theme Theme) {
	// 倒數數字置於右上角圓形進度圈中心
	sec := int((s.TimeLeftMs + 999) / 1000)
	label := fmt.Sprintf("%ds", sec)
	col := theme.Text
	if sec%2 == 0 {
		col = theme.Accent
	}
	w := s.width
	cx := float32(originX + w - 18)
	cy := float32(originY + 18)
	r := float32(16)
	// 先畫圓弧（避免覆蓋文字）
	seg := 36
	end := int(float32(seg) * s.Progress)
	var px, py float32
	for i := 0; i <= end; i++ {
		ang := -float32(math.Pi/2) + 2*float32(math.Pi)*float32(i)/float32(seg)
		x := cx + r*float32(math.Cos(float64(ang)))
		y := cy + r*float32(math.Sin(float64(ang)))
		if i > 0 {
			vector.StrokeLine(dst, px, py, x, y, 2, theme.Accent, true)
		}
		px, py = x, y
	}
	// 再畫置中的倒數文字（在弧線上方），自動縮放以適配圓內
	xf := text.NewGoXFace(face)
	m := xf.Metrics()
	tw, _ := text.Measure(label, xf, 0)
	lineH := m.HAscent + m.HDescent
	// 最大可用寬/高：使用圓的內徑（留 4px 邊距）
	innerD := float64(2*r) - 4
	if innerD < 6 {
		innerD = 6
	}
	maxW := innerD
	maxH := innerD
	scale := 1.0
	if float64(tw) > maxW || lineH > maxH {
		sw := maxW / float64(tw)
		sh := maxH / lineH
		if sw < sh {
			scale = sw
		} else {
			scale = sh
		}
		if scale < 0.7 { // 下限，避免過小難辨識
			scale = 0.7
		}
	}
	var op text.DrawOptions
	op.GeoM.Scale(scale, scale)
	// 以 baseline 對齊將文字垂直置中到 cy；水平置中到 cx
	bx := float64(cx) - (float64(tw) * scale / 2)
	by := float64(cy) + (m.HAscent*scale - (lineH*scale)/2)
	// 視覺微調：大多數數字/字元沒有真正使用到全部 HDescent，高度會顯得偏下
	// 因此以上移 HDescent 的一部分來校正
	by -= m.HDescent * scale * 0.45
	op.GeoM.Translate(bx, by)
	op.ColorScale.ScaleWithColor(col)
	text.Draw(dst, label, xf, &op)
}
