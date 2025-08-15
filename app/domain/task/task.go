package task

import "time"

type Type string

const (
	Practice Type = "Practice"
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
