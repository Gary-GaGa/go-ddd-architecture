package task

import (
	"time"

	"github.com/hajimehoshi/ebiten/v2"
)

// Renderer 是任務動畫的渲染器介面，由 HUD 調用。
// 契約：
// - Sync 以 VM 的 total/remaining 為唯一真實來源；rewardK 供完成時視覺用。
// - Update 僅推進本地微動畫（spawn、抖動等），不可自行推進進度。
// - Draw 在給定矩形範圍內繪製，使用 Theme 色彩，不與 HUD 其他 UI 重疊。
// - Renderer 不修改 VM 狀態。
//
// 注意：保持每幀零/極少配置，確保 60 FPS。
type Renderer interface {
	Sync(totalSec, remainingSec int64, rewardK int64)
	Update(dt time.Duration)
	Draw(dst *ebiten.Image, x, y, w, h int, theme Theme)
}
