package http

import (
	"encoding/json"
	"net/http"

	"game-server/internal/error_handling"
	"game-server/internal/server/domain"
)

func (h *Handler) CreatePlayer(w http.ResponseWriter, r *http.Request) {
	var req domain.CreatePlayerRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		error_handling.WriteError(w, error_handling.BadRequestError("Invalid request body"))
		return
	}

	player, appErr := h.Services.Player.CreatePlayer(r.Context(), req.Nickname)
	if appErr != nil {
		error_handling.WriteError(w, appErr)
		return
	}

	error_handling.WriteJSON(w, http.StatusCreated, player)
}

func (h *Handler) GetPlayer(w http.ResponseWriter, r *http.Request) {
	playerID := r.URL.Query().Get("playerId")

	player, appErr := h.Services.Player.GetPlayer(r.Context(), playerID)
	if appErr != nil {
		error_handling.WriteError(w, appErr)
		return
	}

	error_handling.WriteJSON(w, http.StatusOK, player)
}
