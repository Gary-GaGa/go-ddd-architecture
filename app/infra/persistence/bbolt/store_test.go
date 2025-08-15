package bbolt

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"go-ddd-architecture/app/domain/gametime"
	"go-ddd-architecture/app/domain/player"
)

func tmpDB(t *testing.T) string {
	t.Helper()
	dir := t.TempDir()
	return filepath.Join(dir, "game.db")
}

// Test first-run behavior: Load on empty DB should not error and should set timestamps to now when zero.
func TestStore_FirstRun_LoadInitializesTimestamps(t *testing.T) {
	path := tmpDB(t)
	s, err := New(path)
	if err != nil {
		t.Fatalf("new: %v", err)
	}
	t.Cleanup(func() { _ = s.Close(); _ = os.Remove(path) })

	p, ts, err := s.Load()
	if err != nil {
		t.Fatalf("load: %v", err)
	}
	_ = p // player zero value acceptable on first run
	if ts.WallClockAtClose.IsZero() {
		t.Fatalf("expected timestamps initialized, got zero")
	}
}

// Test roundtrip: Save then Load returns the same data.
func TestStore_Roundtrip_SaveLoad(t *testing.T) {
	path := tmpDB(t)
	s, err := New(path)
	if err != nil {
		t.Fatalf("new: %v", err)
	}
	t.Cleanup(func() { _ = s.Close(); _ = os.Remove(path) })

	p := player.Player{ID: "p1"}
	ts := gametime.Timestamps{WallClockAtClose: time.Unix(1700000000, 0).UTC()}

	if err := s.Save(p, ts); err != nil {
		t.Fatalf("save: %v", err)
	}

	p2, ts2, err := s.Load()
	if err != nil {
		t.Fatalf("load: %v", err)
	}
	if p2.ID != p.ID {
		t.Fatalf("player mismatch: %q != %q", p2.ID, p.ID)
	}
	if !ts2.WallClockAtClose.Equal(ts.WallClockAtClose) {
		t.Fatalf("timestamps mismatch: %v != %v", ts2.WallClockAtClose, ts.WallClockAtClose)
	}
}
