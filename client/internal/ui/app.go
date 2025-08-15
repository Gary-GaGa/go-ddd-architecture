package ui

import (
	"context"
	"image/color"
	"sync/atomic"
	"time"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
	"github.com/hajimehoshi/ebiten/v2/text"
	"golang.org/x/image/font"
	"golang.org/x/image/font/basicfont"

	"go-ddd-architecture/client/internal/api/gameclient"
)

// App 是 Ebiten 遊戲主迴圈的實作（Phase 0）。
// - 定時輪詢 /viewmodel
// - 監聽鍵盤操作呼叫 API
// - DrawHUD 顯示目前狀態

type App struct {
	api    *gameclient.Client
	state  *State
	face   font.Face
	poll   time.Duration
	busy   atomic.Bool // 目前是否有 API 呼叫進行中（避免重複觸發）
	lastVM time.Time   // 最近更新 VM 的時間

	// 自動完成觸發的去重 key（避免相同任務重複觸發）
	autoFinishKey string // 格式: taskID|endsAt
	// 自動 practice 的去重 key，避免在同一週期重複發起
	autoPracticeKey string

	// 去抖 Networking 顯示
	netShowSince   time.Time     // 開始請求的時間
	netHideDelay   time.Duration // 請求結束後延遲隱藏時間
	netShowLatency time.Duration // 請求開始延遲顯示的門檻

	// 輕量 Toast 提示（例如升級失敗）
	toastMsg   atomic.Value // string
	toastUntil atomic.Value // time.Time
}

func NewApp(api *gameclient.Client, state *State) *App {
	return &App{
		api:   api,
		state: state,
		face:  basicfont.Face7x13,
		poll:  500 * time.Millisecond,
		// 去抖：開始超過 150ms 才顯示；結束後保留 350ms
		netHideDelay:   350 * time.Millisecond,
		netShowLatency: 150 * time.Millisecond,
	}
}

func (a *App) Update() error {
	// 定期輪詢 viewmodel（非阻塞：若 busy 則跳過此輪）
	if time.Since(a.lastVM) >= a.poll && !a.busy.Load() {
		a.busy.Store(true)
		a.netShowSince = time.Now()
		go func() {
			defer a.busy.Store(false)
			ctx, cancel := context.WithTimeout(context.Background(), 1500*time.Millisecond)
			defer cancel()
			vm, err := a.api.GetViewModel(ctx)
			if err != nil {
				a.state.SetErr(err.Error())
				return
			}
			a.state.SetErr("")
			a.state.SetVM(vm)
			a.lastVM = time.Now()
		}()
	}

	// 鍵盤事件（使用 inpututil 按下觸發一次）
	if inpututil.IsKeyJustPressed(ebiten.KeyP) && !a.busy.Load() {
		a.trigger(func(ctx context.Context) error {
			vm, err := a.api.PostStartPractice(ctx)
			if err == nil {
				a.state.SetVM(vm)
			}
			return err
		})
	}
	// 移除 [F] 手動完成鍵，統一使用自動完成（任務倒數 <= 0 自動呼叫 TryFinish）
	if inpututil.IsKeyJustPressed(ebiten.KeyU) && !a.busy.Load() {
		a.trigger(func(ctx context.Context) error {
			vm, err := a.api.PostUpgradeKnowledge(ctx)
			if err == nil {
				a.state.SetVM(vm)
				return nil
			}
			// 若為型別化 API 錯誤，判斷是否為 not_enough_research
			// 直接用字串判斷，因 server 以 code: not_enough_research 回傳
			if s := err.Error(); s != "" && (s == "not_enough_research: not enough research" || s == "not_enough_research") {
				a.showToast("Not enough Research to upgrade")
				// 不把此錯誤覆蓋到全域錯誤列，避免嚇人
				return nil
			}
			return err
		})
	}
	if inpututil.IsKeyJustPressed(ebiten.KeyC) && !a.busy.Load() {
		a.trigger(func(ctx context.Context) error {
			out, err := a.api.PostClaimOffline(ctx, "")
			if err == nil {
				a.state.SetVM(out.ViewModel)
			}
			return err
		})
	}
	// 語言切換：1->go, 2->py, 3->js（低風險、先提供三種預設）
	if inpututil.IsKeyJustPressed(ebiten.Key1) && !a.busy.Load() {
		a.trigger(func(ctx context.Context) error {
			vm, err := a.api.PostSelectLanguage(ctx, "go")
			if err == nil {
				a.state.SetVM(vm)
				a.showToast("切換到 Go")
			}
			return err
		})
	}
	if inpututil.IsKeyJustPressed(ebiten.Key2) && !a.busy.Load() {
		a.trigger(func(ctx context.Context) error {
			vm, err := a.api.PostSelectLanguage(ctx, "py")
			if err == nil {
				a.state.SetVM(vm)
				a.showToast("切換到 Python")
			}
			return err
		})
	}
	if inpututil.IsKeyJustPressed(ebiten.Key3) && !a.busy.Load() {
		a.trigger(func(ctx context.Context) error {
			vm, err := a.api.PostSelectLanguage(ctx, "js")
			if err == nil {
				a.state.SetVM(vm)
				a.showToast("切換到 JavaScript")
			}
			return err
		})
	}
	// 自動嘗試完成任務：剩餘時間 <= 0 即自動呼叫 TryFinish（不用再按 F）
	if !a.busy.Load() {
		vmSnap, _ := a.state.Snapshot()
		// 自動啟動 practice：若沒有任務時自動發起（自我學習）
		if vmSnap.CurrentTask == nil {
			if a.autoPracticeKey == "" {
				a.autoPracticeKey = "inflight"
				a.trigger(func(ctx context.Context) error {
					vm, err := a.api.PostStartPractice(ctx)
					if err == nil {
						a.state.SetVM(vm)
					}
					return err
				})
			}
		}

		if vmSnap.CurrentTask != nil && vmSnap.CurrentTask.RemainingSeconds <= 0 {
			key := vmSnap.CurrentTask.ID + "|" + vmSnap.CurrentTask.EndsAt
			if key != "" && key != a.autoFinishKey {
				a.autoFinishKey = key
				a.trigger(func(ctx context.Context) error {
					out, err := a.api.PostTryFinish(ctx)
					if err == nil {
						a.state.SetVM(out.ViewModel)
					}
					return err
				})
			}
		}
		// 若沒有任務，清空 key（允許下一個任務再觸發）
		if vmSnap.CurrentTask == nil {
			a.autoFinishKey = ""
			// 若伺服器已回應新任務，清除 practice 去重
			if a.autoPracticeKey != "" {
				// 仍保持去重直到下一輪確認
				// (在下一次沒有任務時會再觸發)
			}
		} else {
			// 當存在任務時可以重置 practice 去重，允許未來重新啟動
			a.autoPracticeKey = ""
		}
	}
	return nil
}

func (a *App) trigger(fn func(ctx context.Context) error) {
	a.busy.Store(true)
	a.netShowSince = time.Now()
	go func() {
		defer a.busy.Store(false)
		ctx, cancel := context.WithTimeout(context.Background(), 1500*time.Millisecond)
		defer cancel()
		if err := fn(ctx); err != nil {
			a.state.SetErr(err.Error())
		} else {
			a.state.SetErr("")
		}
	}()
}

func (a *App) Draw(screen *ebiten.Image) {
	vm, errMsg := a.state.Snapshot()
	// 轉成 HUD VM
	hudVM := VM{
		Knowledge:        vm.Knowledge,
		Research:         vm.Research,
		Level:            vm.Level,
		NextUpgradeCost:  vm.NextUpgradeCost,
		KnowledgePerMin:  vm.KnowledgePerMin,
		ResearchPerMin:   vm.ResearchPerMin,
		EstimatedSuccess: vm.EstimatedSuccess,
	}
	// 傳遞多語言資訊給 HUD（固定順序 go/py/js 若存在）
	hudVM.CurrentLanguage = vm.CurrentLanguage
	if len(vm.Languages) > 0 {
		hudVM.Languages = map[string]LangHUD{}
		if s, ok := vm.Languages["go"]; ok {
			hudVM.Languages["go"] = LangHUD{Code: "go", Knowledge: s.Knowledge, Research: s.Research, Level: s.Level}
		}
		if s, ok := vm.Languages["py"]; ok {
			hudVM.Languages["py"] = LangHUD{Code: "py", Knowledge: s.Knowledge, Research: s.Research, Level: s.Level}
		}
		if s, ok := vm.Languages["js"]; ok {
			hudVM.Languages["js"] = LangHUD{Code: "js", Knowledge: s.Knowledge, Research: s.Research, Level: s.Level}
		}
		// 其餘語言若有也一併加入（非固定清單）
		for k, s := range vm.Languages {
			if _, fixed := hudVM.Languages[k]; fixed {
				continue
			}
			hudVM.Languages[k] = LangHUD{Code: k, Knowledge: s.Knowledge, Research: s.Research, Level: s.Level}
		}
	}
	if vm.CurrentTask != nil {
		hudVM.CurrentTask = &VMTask{
			Type:             vm.CurrentTask.Type,
			RemainingSeconds: vm.CurrentTask.RemainingSeconds,
			Language:         vm.CurrentTask.Language,
			DurationSeconds:  vm.CurrentTask.DurationSeconds,
			BaseReward:       vm.CurrentTask.BaseReward,
		}
	}
	// 去抖後的網路顯示：開始超過門檻才顯示；結束後保留一段時間
	inFlight := a.busy.Load()
	now := time.Now()
	showNetworking := false
	if inFlight {
		showNetworking = now.Sub(a.netShowSince) >= a.netShowLatency
	} else if !a.netShowSince.IsZero() {
		showNetworking = now.Sub(a.netShowSince) <= (a.netShowLatency + a.netHideDelay)
	}
	// 取得 toast 訊息
	toast := a.getToast()
	DrawHUD(screen, a.face, hudVM, errMsg, showNetworking)
	if toast != "" {
		// 在左側卡片下方顯示一條短提示
		// 這裡重用 HUD 的樣式：以 Good 色字顯示（或 Warn 也可）
		// 位置：靠近卡片底部稍微往下
		x := Theme.Pad8 * 2
		y := Theme.Pad8*2 + 320 + Theme.Pad8 + 92 + Theme.Pad8 + 12 // Status 卡 + Task 卡 + padding
		ebitenutilDrawText(screen, a.face, toast, x, y, Theme.Good)
	}

	// 目前移除右側 AI 顯示（暫時不影響遊戲設計）
	// 若日後恢復，請在此重新計算位置並呼叫 DrawEvolvingAI
}

// ebitenutilDrawText 是簡化版文字繪製封裝，避免重複 import text 包（此檔已 import）
// 這裡直接用 text.Draw，作為單行字的便利函式。
func ebitenutilDrawText(dst *ebiten.Image, face font.Face, s string, x, y int, col color.RGBA) {
	text.Draw(dst, s, face, x, y, col)
}

func (a *App) Layout(outsideWidth, outsideHeight int) (screenWidth, screenHeight int) {
	return 960, 540
}

// showToast 顯示一則短暫提示訊息（1.5 秒）。
func (a *App) showToast(msg string) {
	a.toastMsg.Store(msg)
	a.toastUntil.Store(time.Now().Add(1500 * time.Millisecond))
}

// getToast 若仍在顯示期間，回傳訊息；否則回空字串。
func (a *App) getToast() string {
	v := a.toastUntil.Load()
	if v == nil {
		return ""
	}
	until, ok := v.(time.Time)
	if !ok {
		return ""
	}
	if time.Now().After(until) {
		return ""
	}
	if m := a.toastMsg.Load(); m != nil {
		if s, ok := m.(string); ok {
			return s
		}
	}
	return ""
}
