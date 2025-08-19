package game

// ViewModelDto 為最小展示資料，用於 CLI/UI。
type ViewModelDto struct {
	Knowledge       int64
	Research        int64
	Notices         []string
	CurrentTask     *TaskInfo
	Level           int
	NextUpgradeCost int64
	KnowledgePerMin int64
	ResearchPerMin  int64

	// --- Multi-language ---
	CurrentLanguage string
	// Languages 將每個語言的 Knowledge/Research/Level 暴露給前端顯示。
	Languages map[string]LanguageStats
	// 任務成功率預估（0~1），以目前語言計算
	EstimatedSuccess float64

	// --- Store / Hardware ---
	Servers int
	GPUs    int
	Slots   int // total slots = Servers * SlotsPerServer
}

type TaskInfo struct {
	ID   string
	Type string
	// 語言代碼（若適用）
	Language string
	// 任務剩餘秒數（伺服器視角的 now 計算）
	RemainingSeconds int64
	// 結束時間（RFC3339 UTC 字串，便於前端顯示）
	EndsAt string
	// 額外：任務基礎秒數（顯示用途）
	DurationSeconds int64
	// 任務基礎獎勵（顯示用途）
	BaseReward int64
}

type LanguageStats struct {
	Knowledge int64 `json:"knowledge"`
	Research  int64 `json:"research"`
	Level     int   `json:"level"`
}
