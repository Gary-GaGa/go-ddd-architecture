package memory

import (
	"time"

	"go-ddd-architecture/app/domain/gametime"
	"go-ddd-architecture/app/domain/player"
	outPort "go-ddd-architecture/app/usecase/port/out/game"
)

// InMemoryRepo 為簡單記憶體儲存，供測試/展示使用。
type InMemoryRepo struct {
	P  player.Player
	TS gametime.Timestamps
}

func NewInMemoryRepo() *InMemoryRepo {
	return &InMemoryRepo{P: player.Player{}, TS: gametime.Timestamps{WallClockAtClose: time.Now()}}
}

func (r *InMemoryRepo) Load() (player.Player, gametime.Timestamps, error) { return r.P, r.TS, nil }
func (r *InMemoryRepo) Save(p player.Player, ts gametime.Timestamps) error {
	r.P = p
	r.TS = ts
	return nil
}

var _ outPort.Repository = (*InMemoryRepo)(nil)
