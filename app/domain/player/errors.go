package player

import "errors"

// Domain errors — 呼叫端可用 errors.Is() 判斷業務規則違反原因。
var (
	// ErrTaskAlreadyActive 當玩家已有進行中的任務時嘗試啟動新任務。
	ErrTaskAlreadyActive = errors.New("a task is already in progress")

	// ErrInsufficientResearch 研究點不足以執行升級。
	ErrInsufficientResearch = errors.New("not enough research to upgrade")

	// ErrInsufficientKnowledge 知識點不足以購買物品。
	ErrInsufficientKnowledge = errors.New("not enough knowledge")

	// ErrNoSlotAvailable 沒有可用的 GPU 插槽。
	ErrNoSlotAvailable = errors.New("no GPU slot available; buy a server first")
)
