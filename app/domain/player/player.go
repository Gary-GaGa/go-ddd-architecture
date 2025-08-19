package player

import (
	"math"
	"math/rand"
	"time"

	"go-ddd-architecture/app/domain/resource"
	"go-ddd-architecture/app/domain/task"
)

// Player 作為聚合根，聚合錢包與語言進度（MVP 先聚焦錢包與時間）。
type Player struct {
	ID       string
	Wallet   resource.Wallet
	LastSeen time.Time
	Prestige int
	// 全域 Level 保留以維持相容（未來可移除或轉為衍生）。
	Level   int
	Current *task.Task

	// --- Multi-language 擴充 ---
	// CurrentLanguage 目前練習中的語言代碼，例如 "go", "py"。
	CurrentLanguage string
	// Skills 為每個語言的進度。
	Skills map[string]Skill

	// --- 硬體商店（全域，共用）---
	// Servers 提供可安裝顯卡的插槽，GPUs 佔用插槽並提升研究基礎產率。
	Servers int
	GPUs    int
}

// Skill 描述單一語言的熟練度與研究點與等級。
type Skill struct {
	Knowledge int64
	Research  int64
	Level     int
}

// ApplyOfflineGains 應用離線收益的意圖方法。
func (p *Player) ApplyOfflineGains(knowledge, research int64, now time.Time) {
	// 改為各語言各自累計：只加到當前語言（不再同步到全域錢包）。
	if p.CurrentLanguage != "" {
		s := p.ensureSkill(p.CurrentLanguage)
		s.Knowledge += knowledge
		s.Research += research
		p.Skills[p.CurrentLanguage] = s
	}
	p.LastSeen = now
}

// StartPractice 啟動一個固定設定的練習任務（MVP）。
func (p *Player) StartPractice(now time.Time) {
	if p.Current != nil && p.Current.IsActive() {
		return
	}
	lang := p.CurrentLanguage
	if lang == "" {
		lang = "go" // 預設一個語言，避免空值
		p.CurrentLanguage = lang
	}
	s := p.ensureSkill(lang)
	// 以研究點數縮短任務時間：最高 30% 縮短（研究達 1000 時達到上限）
	base := 5 * time.Second
	reduction := 0.3 * math.Min(1.0, float64(s.Research)/1000.0)
	dur := time.Duration(float64(base) * (1.0 - reduction))
	// 以等級略增獎勵
	reward := int64(10 + s.Level*2)
	t := &task.Task{ID: "practice-5s", Type: task.Practice, Language: lang, Duration: dur, BaseReward: reward}
	t.Start(now)
	p.Current = t
}

// StartTargeted 啟動一個針對當前語言的目標任務：略短時長、略高獎勵。
func (p *Player) StartTargeted(now time.Time) {
	if p.Current != nil && p.Current.IsActive() {
		return
	}
	lang := p.CurrentLanguage
	if lang == "" {
		lang = "go"
		p.CurrentLanguage = lang
	}
	s := p.ensureSkill(lang)
	// 以研究點數縮短任務時間：上限 40% 縮短（比 Practice 略高）
	base := 4 * time.Second
	reduction := 0.4 * math.Min(1.0, float64(s.Research)/1200.0)
	dur := time.Duration(float64(base) * (1.0 - reduction))
	// 獎勵略高於 Practice
	reward := int64(12 + s.Level*3)
	t := &task.Task{ID: "targeted-4s", Type: task.Targeted, Language: lang, Duration: dur, BaseReward: reward}
	t.Start(now)
	p.Current = t
}

// StartDeploy 啟動部署任務：中等時長、較高知識獎勵。
func (p *Player) StartDeploy(now time.Time) {
	if p.Current != nil && p.Current.IsActive() {
		return
	}
	lang := p.CurrentLanguage
	if lang == "" {
		lang = "go"
		p.CurrentLanguage = lang
	}
	s := p.ensureSkill(lang)
	// 部署：基礎 5s，最高 35% 縮短（研究達 1100）
	base := 5 * time.Second
	reduction := 0.35 * math.Min(1.0, float64(s.Research)/1100.0)
	dur := time.Duration(float64(base) * (1.0 - reduction))
	reward := int64(14 + s.Level*3)
	t := &task.Task{ID: "deploy-5s", Type: task.Deploy, Language: lang, Duration: dur, BaseReward: reward}
	t.Start(now)
	p.Current = t
}

// StartResearch 啟動研究任務：時長略長、知識獎勵較溫和（研究獎勵沿用既有邏輯）。
func (p *Player) StartResearch(now time.Time) {
	if p.Current != nil && p.Current.IsActive() {
		return
	}
	lang := p.CurrentLanguage
	if lang == "" {
		lang = "go"
		p.CurrentLanguage = lang
	}
	s := p.ensureSkill(lang)
	// 研究：基礎 6s，最高 45% 縮短（研究達 1400）
	base := 6 * time.Second
	reduction := 0.45 * math.Min(1.0, float64(s.Research)/1400.0)
	dur := time.Duration(float64(base) * (1.0 - reduction))
	reward := int64(9 + s.Level*2)
	t := &task.Task{ID: "research-6s", Type: task.Research, Language: lang, Duration: dur, BaseReward: reward}
	t.Start(now)
	p.Current = t
}

// TryFinish 嘗試完成當前任務，若完成則結算獎勵。
func (p *Player) TryFinish(now time.Time) (finished bool, reward int64) {
	if p.Current == nil || !p.Current.IsActive() {
		return false, 0
	}
	if !p.Current.Done(now) {
		return false, 0
	}
	// 成功率：應以任務啟動時的語言為準（Task.Language），避免切換語言造成歸屬錯誤
	reward = p.Current.BaseReward
	taskLang := p.Current.Language
	prob := p.EstimatedSuccessFor(taskLang)
	rng := rand.New(rand.NewSource(now.UnixNano()))
	if rng.Float64() <= prob {
		// 成功：知識/研究回饋到啟動任務時的語言（多語言獨立累計）。
		// Research 獎勵：基礎 0~3，隨顯卡數量將基數平移（例如 1 張顯卡 => 1~4）。
		gainedRes := int64(rng.Intn(4) + p.GPUs) // [GPUs .. GPUs+3]
		if taskLang != "" {
			s := p.ensureSkill(taskLang)
			s.Knowledge += reward
			s.Research += gainedRes
			p.Skills[taskLang] = s
		}
	} else {
		// 失敗：無獎勵，但任務結束。
		reward = 0
	}
	p.Current.Finish()
	p.Current = nil
	return true, reward
}

// EstimatedSuccess 回傳目前語言的任務成功機率（0~1）。
// 基礎 60%，隨 Knowledge 緩增至 95% 封頂；語言係數略做調整。
func (p *Player) EstimatedSuccess() float64 {
	return p.EstimatedSuccessFor(p.CurrentLanguage)
}

// EstimatedSuccessFor 計算指定語言的成功率（0~1）。
// 使用與 EstimatedSuccess 相同的邏輯，但可針對任務啟動時的語言計算。
func (p *Player) EstimatedSuccessFor(lang string) float64 {
	base := 0.60
	inc := 0.0
	if lang != "" {
		// 指數遞減（diminishing returns）
		s := p.ensureSkill(lang)
		K := float64(s.Knowledge)
		maxInc := 0.35
		scale := 400.0
		inc = maxInc * (1.0 - math.Exp(-K/scale))
	}
	prob := base + inc
	// 語言難度係數（微調）
	switch lang {
	case "py", "python":
		prob += 0.02
	case "js", "javascript":
		prob -= 0.03
	}
	if prob < 0.05 {
		prob = 0.05
	} else if prob > 0.98 {
		prob = 0.98
	}
	return prob
}

// KnowledgePerMinute 根據等級計算知識每分鐘產率（MVP 公式）。
func (p *Player) KnowledgePerMinute() int64 {
	// 依語言等級計算，每語言：10 + 2*Level
	level := p.getCurrentLangLevel()
	base := int64(10)
	bonus := int64(level * 2)
	return base + bonus
}

// ResearchPerMinute 根據等級計算研究每分鐘產率（MVP 公式）。
func (p *Player) ResearchPerMinute() int64 {
	// 依語言等級計算，每語言：2 + 1*Level
	level := p.getCurrentLangLevel()
	base := int64(2)
	bonus := int64(level)
	// 顯卡加成：每張顯卡 +GPUBonusRPM
	return base + bonus + int64(p.GPUs*GPUBonusRPM)
}

// NextUpgradeCost 計算下一級所需的研究點數（MVP：100 * 2^Level）。
func (p *Player) NextUpgradeCost() int64 {
	cost := int64(100)
	// 2 的 Level 次方
	level := p.getCurrentLangLevel()
	for i := 0; i < level; i++ {
		cost *= 2
	}
	return cost
}

// UpgradeKnowledge 嘗試進行升級：扣除當前語言的研究點，Level+1（各語言獨立）。
func (p *Player) UpgradeKnowledge() bool {
	cost := p.NextUpgradeCost()
	// 從當前語言的研究點扣款（並維持全域相容：同步扣全域錢包）。
	lang := p.CurrentLanguage
	if lang == "" {
		lang = "go"
		p.CurrentLanguage = lang
	}
	s := p.ensureSkill(lang)
	if s.Research < cost {
		return false
	}
	s.Research -= cost
	s.Level++
	p.Skills[lang] = s
	return true
}

// SelectLanguage 設定當前語言；若尚未存在則初始化技能。
func (p *Player) SelectLanguage(lang string) {
	if lang == "" {
		return
	}
	_ = p.ensureSkill(lang)
	p.CurrentLanguage = lang
}

func (p *Player) ensureSkill(lang string) Skill {
	if p.Skills == nil {
		p.Skills = map[string]Skill{}
	}
	s, ok := p.Skills[lang]
	if !ok {
		s = Skill{Knowledge: 0, Research: 0, Level: 0}
		p.Skills[lang] = s
	}
	return s
}

func (p *Player) getCurrentLangLevel() int {
	if p.CurrentLanguage == "" {
		return p.Level // 相容：若尚未選語言，沿用舊欄位
	}
	if p.Skills == nil {
		return 0
	}
	s, ok := p.Skills[p.CurrentLanguage]
	if !ok {
		return 0
	}
	return s.Level
}

// --- 商店規則與購買邏輯 ---

const (
	SlotsPerServer = 2          // 每台伺服器可安裝顯卡數
	GPUBonusRPM    = 1          // 每張顯卡增加的 Research 每分鐘
	ServerCostK    = int64(150) // 伺服器主機消耗的 Knowledge
	GPUCostK       = int64(80)  // 顯卡消耗的 Knowledge
)

// BuyServer 嘗試購買一台伺服器主機（消耗當前語言的 Knowledge）。
func (p *Player) BuyServer() bool {
	lang := p.CurrentLanguage
	if lang == "" {
		lang = "go"
		p.CurrentLanguage = lang
	}
	s := p.ensureSkill(lang)
	if s.Knowledge < ServerCostK {
		return false
	}
	s.Knowledge -= ServerCostK
	p.Skills[lang] = s
	p.Servers++
	return true
}

// BuyGPU 嘗試購買一張顯卡（需有可用插槽，並消耗 Knowledge）。
func (p *Player) BuyGPU() bool {
	// 容量限制：需先有伺服器以提供插槽
	cap := p.Servers * SlotsPerServer
	if cap <= 0 {
		return false
	}
	if p.GPUs >= cap {
		return false
	}
	lang := p.CurrentLanguage
	if lang == "" {
		lang = "go"
		p.CurrentLanguage = lang
	}
	s := p.ensureSkill(lang)
	if s.Knowledge < GPUCostK {
		return false
	}
	s.Knowledge -= GPUCostK
	p.Skills[lang] = s
	p.GPUs++
	return true
}
