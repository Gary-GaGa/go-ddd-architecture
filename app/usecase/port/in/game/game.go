package game

import (
	"time"

	dto "go-ddd-architecture/app/usecase/dto/game"
)

// Usecase 定義 Game 的 Input Port。
type Usecase interface {
	Initialize() error
	ClaimOffline(now time.Time) (dto.OfflineResultDto, error)
	GetViewModel() dto.ViewModelDto
	StartPractice(now time.Time) error
	StartTargeted(now time.Time) error
	StartDeploy(now time.Time) error
	StartResearch(now time.Time) error
	TryFinish(now time.Time) (finished bool, reward int64, err error)
	UpgradeKnowledge() error
	SelectLanguage(lang string) error
	BuyServer() error
	BuyGPU() error
}
