package gameclient

// 這些型別 mirror 伺服器的 JSON 結構，避免直接 import usecase 的 DTO（降低耦合）。
// 若後端 DTO 調整，本層只需做相容處理即可。

type ViewModel struct {
	Knowledge        int64             `json:"Knowledge"`
	Research         int64             `json:"Research"`
	Notices          []string          `json:"Notices"`
	CurrentTask      *Task             `json:"CurrentTask"`
	Level            int               `json:"Level"`
	NextUpgradeCost  int64             `json:"NextUpgradeCost"`
	KnowledgePerMin  int64             `json:"KnowledgePerMin"`
	ResearchPerMin   int64             `json:"ResearchPerMin"`
	CurrentLanguage  string            `json:"CurrentLanguage"`
	Languages        map[string]LangVM `json:"Languages"`
	EstimatedSuccess float64           `json:"EstimatedSuccess"`
	// Store
	Servers int `json:"Servers"`
	GPUs    int `json:"GPUs"`
	Slots   int `json:"Slots"`
}

type Task struct {
	ID               string `json:"ID"`
	Type             string `json:"Type"`
	Language         string `json:"Language"`
	RemainingSeconds int64  `json:"RemainingSeconds"`
	EndsAt           string `json:"EndsAt"`
	DurationSeconds  int64  `json:"DurationSeconds"`
	BaseReward       int64  `json:"BaseReward"`
}

type LangVM struct {
	Knowledge int64 `json:"knowledge"`
	Research  int64 `json:"research"`
	Level     int   `json:"level"`
}

type ClaimOfflineRequest struct {
	AsOf string `json:"asOf,omitempty"`
}

type ClaimOfflineResult struct {
	DeltaSeconds     int64 `json:"DeltaSeconds"`
	ClampedSeconds   int64 `json:"ClampedSeconds"`
	GainedKnowledge  int64 `json:"GainedKnowledge"`
	GainedResearch   int64 `json:"GainedResearch"`
	ClampedByCeiling bool  `json:"ClampedByCeiling"`
}

type ClaimOfflineResponse struct {
	Result    ClaimOfflineResult `json:"result"`
	ViewModel ViewModel          `json:"viewModel"`
}

type FinishResponse struct {
	Finished  bool      `json:"finished"`
	Reward    int64     `json:"reward"`
	ViewModel ViewModel `json:"viewModel"`
}

type APIError struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

type ErrorEnvelope struct {
	Error APIError `json:"error"`
}
