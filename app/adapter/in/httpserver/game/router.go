package game

import (
	"net/http"
)

type Router struct {
	mux *http.ServeMux
}

func NewRouter(h *Handler) *Router {
	mux := http.NewServeMux()
	mux.HandleFunc("/api/v1/game/viewmodel", h.GetViewModel)
	mux.HandleFunc("/api/v1/game/claim-offline", h.PostClaimOffline)
	mux.HandleFunc("/api/v1/game/start-practice", h.PostStartPractice)
	mux.HandleFunc("/api/v1/game/try-finish", h.PostTryFinish)
	mux.HandleFunc("/api/v1/game/upgrade-knowledge", h.PostUpgradeKnowledge)
	mux.HandleFunc("/api/v1/game/select-language", h.PostSelectLanguage)

	// legacy
	mux.HandleFunc("/api/game/viewmodel", h.GetViewModel)
	mux.HandleFunc("/api/game/claim-offline", h.PostClaimOffline)
	mux.HandleFunc("/api/game/start-practice", h.PostStartPractice)
	mux.HandleFunc("/api/game/try-finish", h.PostTryFinish)
	mux.HandleFunc("/api/game/upgrade-knowledge", h.PostUpgradeKnowledge)
	mux.HandleFunc("/api/game/select-language", h.PostSelectLanguage)
	return &Router{mux: mux}
}

func (r *Router) Handler() http.Handler { return r.mux }
