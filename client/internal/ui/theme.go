package ui

import "image/color"

// Theme：Phase 0.5 的色票與間距常數
var Theme = struct {
	Bg          color.RGBA
	CardBg      color.RGBA
	CardBorder  color.RGBA
	OutlineBlue color.RGBA
	TextMain    color.RGBA
	TextSub     color.RGBA
	Accent      color.RGBA
	Warn        color.RGBA
	Error       color.RGBA
	Good        color.RGBA

	Pad8    int
	Radius8 float32
}{
	Bg:         color.RGBA{0x0B, 0x0E, 0x14, 0xFF},
	CardBg:     color.RGBA{0x14, 0x19, 0x22, 0xE6},
	CardBorder: color.RGBA{0x23, 0x2B, 0x3A, 0xFF},
	// 淡藍色描邊（比 Accent 柔和一些）
	OutlineBlue: color.RGBA{0x58, 0xB8, 0xFF, 0xCC},
	TextMain:    color.RGBA{0xE6, 0xED, 0xF3, 0xFF},
	TextSub:     color.RGBA{0x94, 0xA3, 0xB8, 0xFF},
	Accent:      color.RGBA{0x3D, 0xA9, 0xFC, 0xFF},
	Warn:        color.RGBA{0xF5, 0x9E, 0x0B, 0xFF},
	Error:       color.RGBA{0xF4, 0x3F, 0x5E, 0xFF},
	Good:        color.RGBA{0x5E, 0xEA, 0xD4, 0xFF},

	Pad8:    8,
	Radius8: 8,
}
