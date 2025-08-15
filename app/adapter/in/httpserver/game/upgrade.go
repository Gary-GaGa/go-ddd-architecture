package game

import (
	"net/http"
)

func (h *Handler) PostUpgradeKnowledge(w http.ResponseWriter, r *http.Request) {
	ok, err := h.uc.UpgradeKnowledge()
	if err != nil {
		writeError(w, http.StatusInternalServerError, "internal", err.Error())
		return
	}
	if !ok {
		writeError(w, http.StatusBadRequest, "not_enough_research", "not enough research to upgrade")
		return
	}
	writeJSON(w, http.StatusOK, h.uc.GetViewModel())
}
