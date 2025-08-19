package ui

import (
	"context"
	"fmt"
	"image/color"
	"math/rand"
	"os"
	"path/filepath"
	"sync/atomic"
	"time"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
	"github.com/hajimehoshi/ebiten/v2/vector"
	"golang.org/x/image/font"
	"golang.org/x/image/font/basicfont"
	"golang.org/x/image/font/opentype"

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

	// 升級成功的短暫脈動效果：記錄語言代碼 -> 截止時間
	langPulses map[string]time.Time
	pulseDur   time.Duration

	// 任務完成後的簡短戰鬥動畫（成功：砍半；失敗：逃跑）
	battleType  string    // "success" 或 "fail"
	battleUntil time.Time // 顯示到期時間
	battleLang  string    // which language style to use for battle effects
	battleStyle int       // 1..N：不同式樣

	// 任務完成獎勵的浮出台詞（+Knowledge）
	floatText  string
	floatUntil time.Time
	floatStart time.Time

	// 資源變化微動畫（Status 區 K/R 跳動與飄字）
	prevK int64
	prevR int64
	// 本地跳動/飄字狀態
	kBounceStart time.Time
	kBounceUntil time.Time
	rBounceStart time.Time
	rBounceUntil time.Time
	kFloatText   string
	kFloatStart  time.Time
	kFloatUntil  time.Time
	rFloatText   string
	rFloatStart  time.Time
	rFloatUntil  time.Time

	// Store 安裝動畫
	prevServers     int
	prevGPUs        int
	serverAnimStart time.Time
	serverAnimUntil time.Time
	gpuAnimStart    time.Time
	gpuAnimUntil    time.Time

	// First-time tutorial overlay (dismissable)
	showTutorial bool // 首次教學 Overlay（可略過）
	showHotkeys  bool

	// 語言排序模式與最近使用紀錄
	// langSort: "lv" | "k" | "recent"
	langSort        string
	lastLangUsedAt  map[string]time.Time
	prevCurrentLang string

	// 本地 Task 視覺預覽（不影響後端）："deploy" | "research" | ""
	taskPreview string

	// 當前任務進行中時的排隊意圖："deploy" | "research"
	nextTaskIntent string
}

func NewApp(api *gameclient.Client, state *State) *App {
	return &App{
		api:            api,
		state:          state,
		face:           loadUIFont(),
		poll:           250 * time.Millisecond,
		pulseDur:       300 * time.Millisecond,
		netHideDelay:   300 * time.Millisecond,
		netShowLatency: 120 * time.Millisecond,
		showTutorial:   true,
		showHotkeys:    true,
		langSort:       "lv",
		lastLangUsedAt: map[string]time.Time{},
	}
}

// Update handles periodic polling and input.
func (a *App) Update() error {
	// Periodic viewmodel polling (non-blocking)
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

	// Tutorial overlay keys only
	if a.showTutorial {
		if inpututil.IsKeyJustPressed(ebiten.KeySpace) || inpututil.IsKeyJustPressed(ebiten.KeyEnter) || inpututil.IsKeyJustPressed(ebiten.KeyEscape) {
			a.showTutorial = false
		}
		if inpututil.IsKeyJustPressed(ebiten.KeyH) {
			a.showHotkeys = !a.showHotkeys
		}

		// 本地任務視覺預覽切換（不呼叫後端）：Y 鍵在 practice(預設) → deploy → research → 關閉預覽 間循環
		if inpututil.IsKeyJustPressed(ebiten.KeyY) {
			switch a.taskPreview {
			case "":
				a.taskPreview = "deploy"
				a.showToast("Preview: deploy")
			case "deploy":
				a.taskPreview = "research"
				a.showToast("Preview: research")
			default:
				a.taskPreview = ""
				a.showToast("Preview: off")
			}
		}
		return nil
	}

	// Toggle hotkeys
	if inpututil.IsKeyJustPressed(ebiten.KeyH) {
		a.showHotkeys = !a.showHotkeys
	}

	// 語言排序快捷鍵：F 循環；V=Lv、K=Knowledge、R=Recent
	if inpututil.IsKeyJustPressed(ebiten.KeyF) {
		switch a.langSort {
		case "lv":
			a.langSort = "k"
		case "k":
			a.langSort = "recent"
		default:
			a.langSort = "lv"
		}
		a.showToast("Sort: " + a.langSort)
	}
	if inpututil.IsKeyJustPressed(ebiten.KeyV) {
		a.langSort = "lv"
		a.showToast("Sort: lv")
	}
	if inpututil.IsKeyJustPressed(ebiten.KeyK) {
		a.langSort = "k"
		a.showToast("Sort: k")
	}
	// 移除 R 做為排序快捷鍵，避免與 Research 任務熱鍵衝突。

	// Keyboard actions
	if inpututil.IsKeyJustPressed(ebiten.KeyP) && !a.busy.Load() {
		a.trigger(func(ctx context.Context) error {
			vm, err := a.api.PostStartPractice(ctx)
			if err == nil {
				a.state.SetVM(vm)
			}
			return err
		})
	}
	if inpututil.IsKeyJustPressed(ebiten.KeyT) && !a.busy.Load() {
		a.trigger(func(ctx context.Context) error {
			vm, err := a.api.PostStartTargeted(ctx)
			if err == nil {
				a.state.SetVM(vm)
			}
			return err
		})
	}

	// D: Start Deploy, R: Start Research
	if inpututil.IsKeyJustPressed(ebiten.KeyD) && !a.busy.Load() {
		vmSnap, _ := a.state.Snapshot()
		if vmSnap.CurrentTask != nil && vmSnap.CurrentTask.RemainingSeconds > 0 {
			a.nextTaskIntent = "deploy"
			a.showToast("Queued: Deploy (will start after current task)")
		} else {
			a.trigger(func(ctx context.Context) error {
				vm, err := a.api.PostStartDeploy(ctx)
				if err == nil {
					a.state.SetVM(vm)
					a.showToast("Started Deploy")
				}
				return err
			})
		}
	}
	if inpututil.IsKeyJustPressed(ebiten.KeyR) && !a.busy.Load() {
		vmSnap, _ := a.state.Snapshot()
		if vmSnap.CurrentTask != nil && vmSnap.CurrentTask.RemainingSeconds > 0 {
			a.nextTaskIntent = "research"
			a.showToast("Queued: Research (will start after current task)")
		} else {
			a.trigger(func(ctx context.Context) error {
				vm, err := a.api.PostStartResearch(ctx)
				if err == nil {
					a.state.SetVM(vm)
					a.showToast("Started Research")
				}
				return err
			})
		}
	}

	// Mouse click actions for Store Buy buttons
	if inpututil.IsMouseButtonJustPressed(ebiten.MouseButtonLeft) && !a.busy.Load() {
		mx, my := ebiten.CursorPosition()
		// Server Buy
		if mx >= lastServerBuyRect.X && my >= lastServerBuyRect.Y && mx < lastServerBuyRect.X+lastServerBuyRect.W && my < lastServerBuyRect.Y+lastServerBuyRect.H {
			a.trigger(func(ctx context.Context) error {
				vm, err := a.api.PostBuyServer(ctx)
				if err == nil {
					a.state.SetVM(vm)
					a.showToast("Purchased: Server (+slots)")
					return nil
				}
				a.showToast("Need more Knowledge for Server")
				return err
			})
		}
		// GPU Buy
		if mx >= lastGPUBuyRect.X && my >= lastGPUBuyRect.Y && mx < lastGPUBuyRect.X+lastGPUBuyRect.W && my < lastGPUBuyRect.Y+lastGPUBuyRect.H {
			a.trigger(func(ctx context.Context) error {
				vm, err := a.api.PostBuyGPU(ctx)
				if err == nil {
					a.state.SetVM(vm)
					a.showToast("Purchased: GPU (+Research/min)")
					return nil
				}
				a.showToast("Need slot or Knowledge for GPU")
				return err
			})
		}
	}
	if inpututil.IsKeyJustPressed(ebiten.KeyU) && !a.busy.Load() {
		prev, _ := a.state.Snapshot()
		a.trigger(func(ctx context.Context) error {
			vm, err := a.api.PostUpgradeKnowledge(ctx)
			if err == nil {
				a.state.SetVM(vm)
				now := time.Now()
				for code, after := range vm.Languages {
					if before, ok := prev.Languages[code]; ok {
						if after.Level > before.Level {
							if a.langPulses == nil {
								a.langPulses = map[string]time.Time{}
							}
							a.langPulses[code] = now.Add(a.pulseDur)
						}
					}
				}
				return nil
			}
			if s := err.Error(); s != "" && (s == "not_enough_research: not enough research" || s == "not_enough_research") {
				a.showToast("Not enough Research to upgrade")
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
	// Language switching
	if inpututil.IsKeyJustPressed(ebiten.Key1) && !a.busy.Load() {
		a.trigger(func(ctx context.Context) error {
			vm, err := a.api.PostSelectLanguage(ctx, "go")
			if err == nil {
				a.state.SetVM(vm)
				a.showToast("Switched to Go")
			}
			return err
		})
	}
	if inpututil.IsKeyJustPressed(ebiten.Key2) && !a.busy.Load() {
		a.trigger(func(ctx context.Context) error {
			vm, err := a.api.PostSelectLanguage(ctx, "py")
			if err == nil {
				a.state.SetVM(vm)
				a.showToast("Switched to Python")
			}
			return err
		})
	}
	if inpututil.IsKeyJustPressed(ebiten.Key3) && !a.busy.Load() {
		a.trigger(func(ctx context.Context) error {
			vm, err := a.api.PostSelectLanguage(ctx, "js")
			if err == nil {
				a.state.SetVM(vm)
				a.showToast("Switched to JavaScript")
			}
			return err
		})
	}

	// Auto practice/finish
	if !a.busy.Load() {
		vmSnap, _ := a.state.Snapshot()
		if vmSnap.CurrentTask == nil {
			if a.nextTaskIntent != "" && !a.busy.Load() {
				intent := a.nextTaskIntent
				// 不使用 autoPracticeKey，避免與 practice 邏輯衝突
				a.trigger(func(ctx context.Context) error {
					var (
						vm  gameclient.ViewModel
						err error
					)
					switch intent {
					case "deploy":
						vm, err = a.api.PostStartDeploy(ctx)
					case "research":
						vm, err = a.api.PostStartResearch(ctx)
					default:
						// fallback: practice
						vm, err = a.api.PostStartPractice(ctx)
					}
					if err == nil {
						a.state.SetVM(vm)
						if intent == "deploy" {
							a.showToast("Started Deploy")
						}
						if intent == "research" {
							a.showToast("Started Research")
						}
						a.nextTaskIntent = ""
					}
					return err
				})
			} else if a.autoPracticeKey == "" {
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
						if out.Finished {
							if vmSnap.CurrentTask != nil && vmSnap.CurrentTask.Language != "" {
								a.battleLang = vmSnap.CurrentTask.Language
							} else {
								a.battleLang = out.ViewModel.CurrentLanguage
							}
							r := rand.New(rand.NewSource(time.Now().UnixNano()))
							a.battleStyle = 1 + r.Intn(3)
							if out.Reward > 0 {
								a.battleType = "success"
								a.floatText = fmt.Sprintf("+%d K", out.Reward)
								a.floatStart = time.Now()
								a.floatUntil = a.floatStart.Add(800 * time.Millisecond)
							} else {
								a.battleType = "fail"
							}
							a.battleUntil = time.Now().Add(700 * time.Millisecond)
						}
					}
					return err
				})
			}
		}
		if vmSnap.CurrentTask == nil {
			a.autoFinishKey = ""
			if a.autoPracticeKey != "" {
				// keep until next cycle
			}
		} else {
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
	// 偵測資源變化（只在變大時觸發跳動與飄字）
	nowAnim := time.Now()
	if a.prevK == 0 && vm.Knowledge > 0 {
		a.prevK = vm.Knowledge
	}
	if a.prevR == 0 && vm.Research > 0 {
		a.prevR = vm.Research
	}
	if dk := vm.Knowledge - a.prevK; dk > 0 {
		a.kBounceStart = nowAnim
		a.kBounceUntil = nowAnim.Add(300 * time.Millisecond)
		a.kFloatText = fmt.Sprintf("+%d", dk)
		a.kFloatStart = nowAnim
		a.kFloatUntil = nowAnim.Add(700 * time.Millisecond)
		a.prevK = vm.Knowledge
	} else if vm.Knowledge < a.prevK {
		a.prevK = vm.Knowledge
	}
	if dr := vm.Research - a.prevR; dr > 0 {
		a.rBounceStart = nowAnim
		a.rBounceUntil = nowAnim.Add(300 * time.Millisecond)
		a.rFloatText = fmt.Sprintf("+%d", dr)
		a.rFloatStart = nowAnim
		a.rFloatUntil = nowAnim.Add(700 * time.Millisecond)
		a.prevR = vm.Research
	} else if vm.Research < a.prevR {
		a.prevR = vm.Research
	}
	// 偵測 Store 變化
	if a.prevServers == 0 && vm.Servers > 0 {
		a.prevServers = vm.Servers
	}
	if a.prevGPUs == 0 && vm.GPUs > 0 {
		a.prevGPUs = vm.GPUs
	}
	if vm.Servers > a.prevServers {
		a.serverAnimStart = nowAnim
		a.serverAnimUntil = nowAnim.Add(500 * time.Millisecond)
		a.prevServers = vm.Servers
	} else if vm.Servers < a.prevServers {
		a.prevServers = vm.Servers
	}
	if vm.GPUs > a.prevGPUs {
		a.gpuAnimStart = nowAnim
		a.gpuAnimUntil = nowAnim.Add(500 * time.Millisecond)
		a.prevGPUs = vm.GPUs
	} else if vm.GPUs < a.prevGPUs {
		a.prevGPUs = vm.GPUs
	}
	// 轉成 HUD VM
	hudVM := VM{
		Knowledge:        vm.Knowledge,
		Research:         vm.Research,
		Level:            vm.Level,
		NextUpgradeCost:  vm.NextUpgradeCost,
		KnowledgePerMin:  vm.KnowledgePerMin,
		ResearchPerMin:   vm.ResearchPerMin,
		EstimatedSuccess: vm.EstimatedSuccess,
		Servers:          vm.Servers,
		GPUs:             vm.GPUs,
		Slots:            vm.Slots,
		ShowHotkeys:      a.showHotkeys,
		LangSort:         a.langSort,
		TaskPreview:      a.taskPreview,
	}
	// 將動畫狀態灌入 HUD VM
	hudVM.KBounceStart = a.kBounceStart
	hudVM.KBounceUntil = a.kBounceUntil
	hudVM.RBounceStart = a.rBounceStart
	hudVM.RBounceUntil = a.rBounceUntil
	hudVM.KFloatText = a.kFloatText
	hudVM.KFloatStart = a.kFloatStart
	hudVM.KFloatUntil = a.kFloatUntil
	hudVM.RFloatText = a.rFloatText
	hudVM.RFloatStart = a.rFloatStart
	hudVM.RFloatUntil = a.rFloatUntil
	hudVM.ServerAnimStart = a.serverAnimStart
	hudVM.ServerAnimUntil = a.serverAnimUntil
	hudVM.GPUAnimStart = a.gpuAnimStart
	hudVM.GPUAnimUntil = a.gpuAnimUntil
	// 傳遞多語言資訊給 HUD（固定順序 go/py/js 若存在）
	hudVM.CurrentLanguage = vm.CurrentLanguage
	// 更新最近使用語言時間（偵測變更）
	if vm.CurrentLanguage != "" && vm.CurrentLanguage != a.prevCurrentLang {
		a.lastLangUsedAt[vm.CurrentLanguage] = time.Now()
		a.prevCurrentLang = vm.CurrentLanguage
	}
	// 傳遞最近使用的時間戳（unix 秒）供 HUD 排序
	if len(a.lastLangUsedAt) > 0 {
		hudVM.LangRecent = map[string]int64{}
		for code, t := range a.lastLangUsedAt {
			hudVM.LangRecent[code] = t.Unix()
		}
	}
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
	// 清理過期脈動並傳遞給 HUD
	now2 := time.Now()
	if len(a.langPulses) > 0 {
		for k, until := range a.langPulses {
			if now2.After(until) {
				delete(a.langPulses, k)
			}
		}
		if len(a.langPulses) > 0 {
			// 複製一份傳遞，避免資料競態
			p := make(map[string]time.Time, len(a.langPulses))
			for k, v := range a.langPulses {
				p[k] = v
			}
			hudVM.Pulses = p
		}
	}
	// 戰鬥動畫狀態傳遞（僅在有效期間內傳給 HUD）
	if !a.battleUntil.IsZero() && time.Now().Before(a.battleUntil) {
		hudVM.BattleType = a.battleType
		hudVM.BattleUntil = a.battleUntil
		hudVM.BattleLang = a.battleLang
		hudVM.BattleStyle = a.battleStyle
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
		// 使用 HUD 的封裝繪字（含 baseline 調整）
		drawText(screen, a.face, toast, x, y, Theme.Good)
	}

	// 浮出 +Knowledge 數字（位於畫面中央，往上飄並淡出）
	if a.floatText != "" && !a.floatUntil.IsZero() {
		now := time.Now()
		if now.Before(a.floatUntil) {
			dur := a.floatUntil.Sub(a.floatStart).Seconds()
			t := now.Sub(a.floatStart).Seconds() / dur
			if t < 0 {
				t = 0
			}
			if t > 1 {
				t = 1
			}
			sw, sh := screen.Size()
			y0 := sh/2 + 10
			y1 := sh/2 - 30
			y := int(float64(y0) + (float64(y1-y0) * t))
			alpha := uint8(255 * (1.0 - t))
			col := color.RGBA{R: Theme.Accent.R, G: Theme.Accent.G, B: Theme.Accent.B, A: alpha}
			x := (sw - textWidth(a.face, a.floatText)) / 2
			drawText(screen, a.face, a.floatText, x, y, col)
		} else {
			a.floatText = ""
		}
	}

	// 目前移除右側 AI 顯示（暫時不影響遊戲設計）
	// 若日後恢復，請在此重新計算位置並呼叫 DrawEvolvingAI

	// Tutorial overlay
	if a.showTutorial {
		sw, sh := screen.Size()
		overlay := color.RGBA{0, 0, 0, 140}
		vector.DrawFilledRect(screen, 0, 0, float32(sw), float32(sh), overlay, true)
		w := min(sw-120, 640)
		h := 220
		x := (sw - w) / 2
		y := (sh - h) / 2
		drawRoundedFilledRect(screen, x, y, w, h, 8, Theme.CardBg)
		drawRoundedRectOutline(screen, x, y, w, h, 8, Theme.OutlineBlue, 1)
		tx := x + Theme.Pad8
		ty := y + Theme.Pad8 + 12
		drawText(screen, a.face, "Quick Tour", tx, ty, Theme.TextMain)
		ty += 16
		vector.StrokeLine(screen, float32(tx), float32(ty), float32(x+w-Theme.Pad8), float32(ty), 1, Theme.CardBorder, true)
		ty += 12
		lines := []string{
			"Welcome! This is a tiny practice-and-unlock game.",
			"Left Status shows resources and level; center Task starts/waits tasks.",
			"Use keys: P practice, T targeted, D deploy, R research; U upgrade; C claim.",
			"Click Buy to purchase hardware (Server/GPU).",
		}
		for _, s := range lines {
			drawText(screen, a.face, s, tx, ty, Theme.TextSub)
			ty += 18
		}
		hint := "Press Space / Enter / Esc to dismiss"
		drawText(screen, a.face, hint, x+w-textWidth(a.face, hint)-Theme.Pad8, y+h-Theme.Pad8, Theme.TextSub)
	}
}

// 移除舊版 ebitenutilDrawText，統一改用 text/v2 DrawOptions 於呼叫點處理。

func (a *App) Layout(outsideWidth, outsideHeight int) (screenWidth, screenHeight int) {
	return 960, 540
}

// loadUIFont 嘗試載入支援中文的字型（優先 APP_UI_FONT 指定的檔案），找不到時退回 basicfont。
func loadUIFont() font.Face {
	// 1) 環境變數指定檔案路徑
	if p := os.Getenv("APP_UI_FONT"); p != "" {
		if f := openTTFFace(p, 14); f != nil {
			return f
		}
	}
	// 2) 專案常見路徑候選
	candidates := []string{
		"assets/fonts/NotoSansTC-Regular.ttf",
		"assets/fonts/NotoSansSC-Regular.ttf",
		"assets/fonts/SourceHanSansTC-Regular.otf",
		"assets/fonts/SourceHanSans-Regular.otf",
		"client/assets/fonts/NotoSansTC-Regular.ttf",
		"client/assets/fonts/NotoSansSC-Regular.ttf",
		"client/assets/fonts/SourceHanSansTC-Regular.otf",
		"client/assets/fonts/SourceHanSans-Regular.otf",
	}
	cwd, _ := os.Getwd()
	for _, rel := range candidates {
		p := filepath.Join(cwd, rel)
		if f := openTTFFace(p, 14); f != nil {
			return f
		}
	}
	// 3) 退回內建等寬字：不支援中文，但至少不會崩潰
	return basicfont.Face7x13
}

// openTTFFace 讀取 TTF/OTF 字型，回傳 Face；失敗則回 nil。
func openTTFFace(path string, size float64) font.Face {
	data, err := os.ReadFile(path)
	if err != nil || len(data) == 0 {
		return nil
	}
	ft, err := opentype.Parse(data)
	if err != nil {
		return nil
	}
	face, err := opentype.NewFace(ft, &opentype.FaceOptions{Size: size, DPI: 72})
	if err != nil {
		return nil
	}
	return face
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
