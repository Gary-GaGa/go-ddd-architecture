package resource

// Wallet 表示玩家的資源錢包。
// 使用 int64 以避免溢位風險並方便持久化。
type Wallet struct {
	Knowledge int64
	Research  int64
}

// Add 以原子意圖新增資源（不允許負數結果）。
func (w *Wallet) Add(knowledgeDelta, researchDelta int64) {
	w.Knowledge += knowledgeDelta
	if w.Knowledge < 0 {
		w.Knowledge = 0
	}

	w.Research += researchDelta
	if w.Research < 0 {
		w.Research = 0
	}
}
