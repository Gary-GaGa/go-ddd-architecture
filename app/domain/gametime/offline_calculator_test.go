package gametime

import (
	"testing"
	"time"

	"go-ddd-architecture/app/domain/player"
)

func TestOfflineCalculator_HappyPath(t *testing.T) {
	calc := NewOfflineCalculator()
	p := &player.Player{CurrentLanguage: "go"}

	closeAt := time.Date(2025, 8, 10, 10, 0, 0, 0, time.UTC)
	now := closeAt.Add(30 * time.Minute)
	res := calc.Compute(p, Timestamps{WallClockAtClose: closeAt}, now)

	if res.GainedKnowledge <= 0 || res.GainedResearch <= 0 {
		t.Fatalf("expected positive gains, got %+v", res)
	}
	s := p.Skills["go"]
	if s.Knowledge != res.GainedKnowledge || s.Research != res.GainedResearch {
		t.Fatalf("lang not applied, got skill=%+v res=%+v", s, res)
	}
}

func TestOfflineCalculator_Clamp8h(t *testing.T) {
	calc := NewOfflineCalculator()
	p := &player.Player{CurrentLanguage: "go"}

	closeAt := time.Date(2025, 8, 10, 0, 0, 0, 0, time.UTC)
	now := closeAt.Add(12 * time.Hour)
	res := calc.Compute(p, Timestamps{WallClockAtClose: closeAt}, now)

	if !res.ClampedTo8h {
		t.Fatalf("expected clamped true")
	}
	minutes := int64((8 * time.Hour) / time.Minute)
	expectedK := minutes * 10
	if p.Skills["go"].Knowledge != expectedK {
		t.Fatalf("expected knowledge=%d got %d", expectedK, p.Skills["go"].Knowledge)
	}
}

func TestOfflineCalculator_NegativeTime(t *testing.T) {
	calc := NewOfflineCalculator()
	p := &player.Player{CurrentLanguage: "go"}

	now := time.Date(2025, 8, 10, 10, 0, 0, 0, time.UTC)
	closeAt := now.Add(10 * time.Minute)
	res := calc.Compute(p, Timestamps{WallClockAtClose: closeAt}, now)

	if !res.AnomalyDetected {
		t.Fatalf("expected anomaly detected")
	}
	if s, ok := p.Skills["go"]; ok {
		if s.Knowledge != 0 || s.Research != 0 {
			t.Fatalf("expected no gains, got %+v", s)
		}
	}
}
