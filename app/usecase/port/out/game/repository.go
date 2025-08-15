package game

import (
	"go-ddd-architecture/app/domain/gametime"
	"go-ddd-architecture/app/domain/player"
)

// Repository 定義 Game 用例的持久化 Port。
type Repository interface {
	Load() (player.Player, gametime.Timestamps, error)
	Save(player.Player, gametime.Timestamps) error
}
