package ui

import (
	"sync"

	"go-ddd-architecture/client/internal/api/gameclient"
)

// State 保存 UI 需要的資料，並提供 thread-safe 更新。
type State struct {
	mu sync.RWMutex

	VM gameclient.ViewModel

	// 錯誤訊息（例如 API 失敗）
	Err string
}

func (s *State) SetVM(vm gameclient.ViewModel) {
	s.mu.Lock()
	s.VM = vm
	s.mu.Unlock()
}

func (s *State) SetErr(err string) {
	s.mu.Lock()
	s.Err = err
	s.mu.Unlock()
}

func (s *State) Snapshot() (vm gameclient.ViewModel, errMsg string) {
	s.mu.RLock()
	vm = s.VM
	errMsg = s.Err
	s.mu.RUnlock()
	return
}
