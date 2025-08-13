// Package main - CLI MVP for AI Pet Idle Game (single-file version)
package main

import (
	"bufio"
	"fmt"
	"os"
	"strings"
	"time"
)

// ===== Clock =====

type Clock interface{ Now() time.Time }

type SystemClock struct{}

func (SystemClock) Now() time.Time { return time.Now() }

// ===== Language progression =====

type Language string

const (
	LangGo     Language = "Go"
	LangPython Language = "Python"
)

type LangProgress struct {
	Lang  Language
	Level int
	XP    float64
}

func (p *LangProgress) GainXP(x float64) (leveled bool) {
	p.XP += x
	need := xpNeed(p.Level)
	for p.XP >= need {
		p.XP -= need
		p.Level++
		leveled = true
		need = xpNeed(p.Level)
	}
	return
}

// MVP XP curve: 50 * 1.3^(level-1)
func xpNeed(level int) float64 {
	if level < 1 {
		level = 1
	}
	need := 50.0
	for i := 1; i < level; i++ {
		need *= 1.3
	}
	return need
}

// Each level adds +15% production
func levelMultiplier(level int) float64 { return 1.0 + 0.15*float64(level) }

// ===== Task =====

type TaskType string

const (
	TaskPractice TaskType = "Practice"
	TaskSolve    TaskType = "Solve"
)

type Task struct {
	ID         string
	Type       TaskType
	Duration   time.Duration
	BaseReward int64 // base knowledge reward
	startedAt  time.Time
	doneAt     time.Time
	active     bool
}

func (t *Task) Start(at time.Time) {
	if t.active {
		return
	}
	t.startedAt = at
	t.doneAt = at.Add(t.Duration)
	t.active = true
}
func (t *Task) IsActive() bool         { return t.active }
func (t *Task) Done(at time.Time) bool { return t.active && !at.Before(t.doneAt) }
func (t *Task) Finish()                { t.active = false }

// ===== Player aggregate =====

type Wallet struct {
	Knowledge int64
	Research  int64
}

func (w *Wallet) Add(know, res int64) { w.Knowledge += know; w.Research += res }

type Player struct {
	ID       string
	Wallet   Wallet
	Langs    map[Language]*LangProgress
	Current  *Task
	LastSeen time.Time
	Prestige int
}

func NewPlayer(id string, now time.Time) *Player {
	return &Player{
		ID: id,
		Langs: map[Language]*LangProgress{
			LangGo:     {Lang: LangGo, Level: 1, XP: 0},
			LangPython: {Lang: LangPython, Level: 1, XP: 0},
		},
		LastSeen: now,
	}
}

func (p *Player) StartTask(t *Task, now time.Time) {
	if p.Current != nil && p.Current.IsActive() {
		return
	}
	t.Start(now)
	p.Current = t
}

func (p *Player) TryFinish(now time.Time) (finished bool, reward int64) {
	if p.Current == nil || !p.Current.IsActive() {
		return
	}
	if !p.Current.Done(now) {
		return
	}
	// reward = base * (1 + 0.15*GoLevel) * (1 + 0.02*Prestige)
	langMul := levelMultiplier(p.Langs[LangGo].Level)
	prestigeMul := 1.0 + 0.02*float64(p.Prestige)
	total := float64(p.Current.BaseReward) * langMul * prestigeMul
	p.Wallet.Add(int64(total), 0)
	// gain XP
	p.Langs[LangGo].GainXP(float64(p.Current.BaseReward) * 0.2)
	p.Current.Finish()
	p.Current = nil
	return true, int64(total)
}

// Offline income: cap 8h, simulate as completing a practice chunk every 10s
func (p *Player) ClaimOffline(now time.Time) (seconds int64, gained int64) {
	if now.Before(p.LastSeen) {
		p.LastSeen = now
	}
	delta := now.Sub(p.LastSeen)
	if delta <= 0 {
		return 0, 0
	}
	if delta > 8*time.Hour {
		delta = 8 * time.Hour
	}
	chunks := int(delta / (10 * time.Second))
	var total int64
	for i := 0; i < chunks; i++ {
		base := int64(10)
		langMul := levelMultiplier(p.Langs[LangGo].Level)
		prestigeMul := 1.0 + 0.02*float64(p.Prestige)
		total += int64(float64(base) * langMul * prestigeMul)
		p.Langs[LangGo].GainXP(float64(base) * 0.2)
	}
	p.Wallet.Add(total, 0)
	p.LastSeen = now
	return int64(chunks) * 10, total
}

// ===== Game loop (use-case façade) =====

type Game struct {
	clock  Clock
	player *Player
}

func NewGame(clock Clock, p *Player) *Game { return &Game{clock: clock, player: p} }
func (g *Game) Player() *Player            { return g.player }

func (g *Game) StartPractice() {
	t := &Task{ID: "practice-go-10s", Type: TaskPractice, Duration: 10 * time.Second, BaseReward: 10}
	g.player.StartTask(t, g.clock.Now())
}

func (g *Game) Tick() (finished bool, reward int64)     { return g.player.TryFinish(g.clock.Now()) }
func (g *Game) ClaimOffline() (sec int64, gained int64) { return g.player.ClaimOffline(g.clock.Now()) }

// ===== CLI =====

func main() {
	clock := SystemClock{}
	p := NewPlayer("p1", clock.Now())
	g := NewGame(clock, p)

	// offline settle at start
	if sec, gain := g.ClaimOffline(); sec > 0 {
		fmt.Printf("[OFFLINE] 回放 %ds，獲得 %d 知識點\n", sec, gain)
	}

	fmt.Println("== AI 寵物（CLI MVP）==")
	printStatus(g)
	fmt.Println("指令：start(開始練習)、tick(前進1秒)、status(狀態)、claim(離線結算)、prestige(未實作示意)、quit")

	sc := bufio.NewScanner(os.Stdin)
	for {
		fmt.Print("> ")
		if !sc.Scan() {
			break
		}
		switch strings.TrimSpace(sc.Text()) {
		case "start":
			g.StartPractice()
			fmt.Println("開始練習任務（10 秒）")
		case "tick":
			// 模擬 1 秒過去
			time.Sleep(1 * time.Second)
			if done, r := g.Tick(); done {
				fmt.Printf("任務完成！獲得 %d 知識點\n", r)
			} else {
				fmt.Println("…進行中")
			}
		case "status":
			printStatus(g)
		case "claim":
			sec, gain := g.ClaimOffline()
			fmt.Printf("離線結算 %ds，獲得 %d\n", sec, gain)
		case "prestige":
			fmt.Println("（預留）轉生尚未實作：完成主線後可獲永久加成")
		case "quit", "exit":
			fmt.Println("Bye!")
			return
		default:
			fmt.Println("未知指令")
		}
	}
}

func printStatus(g *Game) {
	p := g.Player()
	fmt.Printf("資源：知識點=%d 研發點=%d | Go等級=%d XP=%.1f | 轉生=%d\n",
		p.Wallet.Knowledge, p.Wallet.Research, p.Langs[LangGo].Level, p.Langs[LangGo].XP, p.Prestige)
	if p.Current != nil && p.Current.IsActive() {
		left := time.Until(p.Current.doneAt)
		fmt.Printf("任務：%s 進行中，剩餘約 %s\n", p.Current.ID, left.Truncate(time.Second))
	} else {
		fmt.Println("任務：無")
	}
}
