package task

import (
	"encoding/json"
	"time"
)

type Type string

const (
	Practice Type = "Practice"
	Targeted Type = "Targeted"
	Deploy   Type = "Deploy"
	Research Type = "Research"
)

type Task struct {
	ID   string
	Type Type
	// Language 本次任務所屬語言（多語言擴充）
	Language   string
	Duration   time.Duration
	BaseReward int64

	startedAt time.Time
	doneAt    time.Time
	active    bool
}

func (t *Task) Start(at time.Time) {
	if t.active {
		return
	}
	t.startedAt = at
	t.doneAt = at.Add(t.Duration)
	t.active = true
}

func (t *Task) IsActive() bool { return t.active }

func (t *Task) Done(at time.Time) bool { return t.active && !at.Before(t.doneAt) }

func (t *Task) Finish() { t.active = false }

func (t *Task) RemainingSeconds(at time.Time) int64 {
	if !t.active {
		return 0
	}
	d := t.doneAt.Sub(at)
	if d <= 0 {
		return 0
	}
	return int64(d / time.Second)
}

// taskJSON 是序列化用的 surrogate struct，讓 unexported 欄位可以持久化。
type taskJSON struct {
	ID         string        `json:"id"`
	Type       Type          `json:"type"`
	Language   string        `json:"language"`
	Duration   time.Duration `json:"duration"`
	BaseReward int64         `json:"baseReward"`
	StartedAt  time.Time     `json:"startedAt"`
	DoneAt     time.Time     `json:"doneAt"`
	Active     bool          `json:"active"`
}

func (t Task) MarshalJSON() ([]byte, error) {
	return json.Marshal(taskJSON{
		ID:         t.ID,
		Type:       t.Type,
		Language:   t.Language,
		Duration:   t.Duration,
		BaseReward: t.BaseReward,
		StartedAt:  t.startedAt,
		DoneAt:     t.doneAt,
		Active:     t.active,
	})
}

func (t *Task) UnmarshalJSON(data []byte) error {
	var j taskJSON
	if err := json.Unmarshal(data, &j); err != nil {
		return err
	}
	t.ID = j.ID
	t.Type = j.Type
	t.Language = j.Language
	t.Duration = j.Duration
	t.BaseReward = j.BaseReward
	t.startedAt = j.StartedAt
	t.doneAt = j.DoneAt
	t.active = j.Active
	return nil
}
