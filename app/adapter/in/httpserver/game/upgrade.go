package game

import (
	"errors"
	"net/http"

	"go-ddd-architecture/app/domain/player"
)

func (h *Handler) PostUpgradeKnowledge(w http.ResponseWriter, r *http.Request) {
	if err := h.uc.UpgradeKnowledge(); err != nil {
		if errors.Is(err, player.ErrInsufficientResearch) {
			writeError(w, http.StatusBadRequest, "not_enough_research", err.Error())
			return
		}
		writeError(w, http.StatusInternalServerError, "internal", err.Error())
		return
	}
	writeJSON(w, http.StatusOK, h.uc.GetViewModel())
}
