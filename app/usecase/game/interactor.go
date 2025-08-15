package game

import (
	"time"

	"go-ddd-architecture/app/domain/gametime"
	"go-ddd-architecture/app/domain/player"
	dto "go-ddd-architecture/app/usecase/dto/game"
	outPort "go-ddd-architecture/app/usecase/port/out/game"
)

// Clock 由外部注入，利於測試。
type Clock interface{ Now() time.Time }

// Interactor 將領域服務與儲存協調起來。
type Interactor struct {
	repo outPort.Repository
	clk  Clock
	calc *gametime.OfflineCalculator

	// 快取狀態（載入於 Initialize）
	p  player.Player
	ts gametime.Timestamps
}

func NewInteractor(repo outPort.Repository, clk Clock, calc *gametime.OfflineCalculator) *Interactor {
	return &Interactor{repo: repo, clk: clk, calc: calc}
}

func (uc *Interactor) Initialize() error {
	p, ts, err := uc.repo.Load()
	if err != nil {
		return err
	}
	// 初始化多語言映射避免 nil map
	if p.Skills == nil {
		p.Skills = map[string]player.Skill{}
	}
	// 設定一個預設語言（相容）：若未選擇，預設 go
	if p.CurrentLanguage == "" {
		p.CurrentLanguage = "go"
	}
	uc.p = p
	uc.ts = ts
	return nil
}

func (uc *Interactor) ClaimOffline(now time.Time) (gametime.OfflineResult, error) {
	res := uc.calc.Compute(&uc.p, uc.ts, now)
	// 更新 timestamps 的關閉時間供下次計算
	uc.ts.WallClockAtClose = now
	if err := uc.repo.Save(uc.p, uc.ts); err != nil {
		return res, err
	}
	return res, nil
}

func (uc *Interactor) GetViewModel() dto.ViewModelDto {
	// 對外顯示 Knowledge/Research 以「當前語言」為主（各語言獨立累計）。
	vm := dto.ViewModelDto{}
	if uc.p.CurrentLanguage != "" {
		if s, ok := uc.p.Skills[uc.p.CurrentLanguage]; ok {
			vm.Knowledge = s.Knowledge
			vm.Research = s.Research
		}
	}
	if uc.p.Current != nil && uc.p.Current.IsActive() {
		now := uc.clk.Now().UTC()
		remaining := uc.p.Current.RemainingSeconds(now)
		endsAt := now.Add(time.Duration(remaining) * time.Second).UTC().Format(time.RFC3339)
		vm.CurrentTask = &dto.TaskInfo{
			ID:               uc.p.Current.ID,
			Type:             string(uc.p.Current.Type),
			Language:         uc.p.Current.Language,
			RemainingSeconds: remaining,
			EndsAt:           endsAt,
			DurationSeconds:  int64(uc.p.Current.Duration / time.Second),
			BaseReward:       uc.p.Current.BaseReward,
		}
	}
	// 擴充：等級/升級資訊（以當前語言的等級呈現）
	lvl := uc.p.Level
	if uc.p.CurrentLanguage != "" {
		if s, ok := uc.p.Skills[uc.p.CurrentLanguage]; ok {
			lvl = s.Level
		} else {
			lvl = 0
		}
	}
	vm.Level = lvl
	vm.NextUpgradeCost = uc.p.NextUpgradeCost()
	vm.KnowledgePerMin = uc.p.KnowledgePerMinute()
	vm.ResearchPerMin = uc.p.ResearchPerMinute()
	vm.EstimatedSuccess = uc.p.EstimatedSuccess()

	// Multi-language: expose languages stats and current selection
	vm.CurrentLanguage = uc.p.CurrentLanguage
	if len(uc.p.Skills) > 0 {
		vm.Languages = map[string]dto.LanguageStats{}
		for k, s := range uc.p.Skills {
			vm.Languages[k] = dto.LanguageStats{Knowledge: s.Knowledge, Research: s.Research, Level: s.Level}
		}
	}
	return vm
}

// StartPractice 啟動練習任務
func (uc *Interactor) StartPractice(now time.Time) error {
	uc.p.StartPractice(now)
	return uc.repo.Save(uc.p, uc.ts)
}

// TryFinish 嘗試完成當前任務
func (uc *Interactor) TryFinish(now time.Time) (finished bool, reward int64, err error) {
	finished, reward = uc.p.TryFinish(now)
	if err = uc.repo.Save(uc.p, uc.ts); err != nil {
		return
	}
	return
}

// UpgradeKnowledge 升級等級，扣除研究。
func (uc *Interactor) UpgradeKnowledge() (ok bool, err error) {
	ok = uc.p.UpgradeKnowledge()
	if !ok {
		return false, nil
	}
	if err = uc.repo.Save(uc.p, uc.ts); err != nil {
		return false, err
	}
	return true, nil
}

// SelectLanguage 設定目前操作的語言
func (uc *Interactor) SelectLanguage(lang string) error {
	uc.p.SelectLanguage(lang)
	return uc.repo.Save(uc.p, uc.ts)
}
