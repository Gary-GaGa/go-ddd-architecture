package httpserver

import (
	"net/http"

	"go.uber.org/zap"

	gameModule "go-ddd-architecture/app/adapter/in/httpserver/game"
)

type Router struct {
	mux *http.ServeMux
}

func NewRouter(gameRouter *gameModule.Router) *Router {
	mux := http.NewServeMux()
	// 掛載 game 模組（它已註冊自身路由於 mux 上）
	// 這裡也可改為子路由分發邏輯，現階段直接使用其 mux
	// 為了統一出口 middleware，我們把子路由的 Handler 掛到一個前綴下
	mux.Handle("/", gameRouter.Handler())
	return &Router{mux: mux}
}

// 包上預設 middleware 鏈
func (r *Router) HandlerWithLogger(log *zap.Logger) http.Handler {
	return Chain(r.mux,
		NewRecoveryMiddleware(log),
		RequestIDMiddleware,
		NewAccessLogMiddleware(log),
	)
}
