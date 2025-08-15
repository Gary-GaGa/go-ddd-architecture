package game

import (
	"encoding/json"
	"net/http"
	"time"

	"go.uber.org/zap"

	"go-ddd-architecture/app/domain/gametime"
	dto "go-ddd-architecture/app/usecase/dto/game"
	inPort "go-ddd-architecture/app/usecase/port/in/game"
)

type Handler struct {
	uc  inPort.Usecase
	log *zap.Logger
}

func NewHandler(uc inPort.Usecase, log *zap.Logger) *Handler { return &Handler{uc: uc, log: log} }

func (h *Handler) GetViewModel(w http.ResponseWriter, r *http.Request) {
	vm := h.uc.GetViewModel()
	writeJSON(w, http.StatusOK, vm)
}

type claimReq struct {
	AsOf string `json:"asOf"`
}

type claimResp struct {
	Result    gametime.OfflineResult `json:"result"`
	ViewModel dto.ViewModelDto       `json:"viewModel"`
}

func (h *Handler) PostClaimOffline(w http.ResponseWriter, r *http.Request) {
	var body claimReq
	if r.Body != nil {
		_ = json.NewDecoder(r.Body).Decode(&body)
	}
	now := time.Now().UTC()
	if body.AsOf != "" {
		if t, err := time.Parse(time.RFC3339, body.AsOf); err == nil {
			now = t
		} else {
			writeError(w, http.StatusBadRequest, "bad_request", "invalid asOf, must be RFC3339")
			return
		}
	}
	res, err := h.uc.ClaimOffline(now)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "internal", err.Error())
		return
	}
	vm := h.uc.GetViewModel()
	writeJSON(w, http.StatusOK, claimResp{Result: res, ViewModel: vm})
}

// Start a practice task immediately
func (h *Handler) PostStartPractice(w http.ResponseWriter, r *http.Request) {
	now := time.Now().UTC()
	if err := h.uc.StartPractice(now); err != nil {
		writeError(w, http.StatusInternalServerError, "internal", err.Error())
		return
	}
	writeJSON(w, http.StatusOK, h.uc.GetViewModel())
}

// Try to finish current task
type finishResp struct {
	Finished  bool             `json:"finished"`
	Reward    int64            `json:"reward"`
	ViewModel dto.ViewModelDto `json:"viewModel"`
}

func (h *Handler) PostTryFinish(w http.ResponseWriter, r *http.Request) {
	now := time.Now().UTC()
	finished, reward, err := h.uc.TryFinish(now)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "internal", err.Error())
		return
	}
	writeJSON(w, http.StatusOK, finishResp{Finished: finished, Reward: reward, ViewModel: h.uc.GetViewModel()})
}

// Select current language
type selectLangReq struct {
	Language string `json:"language"`
}

func (h *Handler) PostSelectLanguage(w http.ResponseWriter, r *http.Request) {
	var body selectLangReq
	if r.Body != nil {
		_ = json.NewDecoder(r.Body).Decode(&body)
	}
	if body.Language == "" {
		writeError(w, http.StatusBadRequest, "bad_request", "language is required")
		return
	}
	if err := h.uc.SelectLanguage(body.Language); err != nil {
		writeError(w, http.StatusInternalServerError, "internal", err.Error())
		return
	}
	writeJSON(w, http.StatusOK, h.uc.GetViewModel())
}

// --- shared helpers (local to game module) ---

type httpError struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

type errorEnvelope struct {
	Err httpError `json:"error"`
}

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}

func writeError(w http.ResponseWriter, status int, code, msg string) {
	writeJSON(w, status, errorEnvelope{Err: httpError{Code: code, Message: msg}})
}
