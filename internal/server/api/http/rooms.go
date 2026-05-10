package http

import (
	"encoding/json"
	"net/http"

	"game-server/internal/error_handling"
	"game-server/internal/server/domain"
)

func (h *Handler) CreateRoom(w http.ResponseWriter, r *http.Request) {
	var req domain.CreateRoomRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		error_handling.WriteError(w, error_handling.BadRequestError("Invalid request body"))
		return
	}

	room, appErr := h.Services.Room.CreateRoom(r.Context(), req.Name, req.MaxPlayers)
	if appErr != nil {
		error_handling.WriteError(w, appErr)
		return
	}

	error_handling.WriteJSON(w, http.StatusCreated, room)
}

func (h *Handler) GetRoom(w http.ResponseWriter, r *http.Request) {
	roomID := r.URL.Query().Get("roomId")

	room, appErr := h.Services.Room.GetRoom(r.Context(), roomID)
	if appErr != nil {
		error_handling.WriteError(w, appErr)
		return
	}

	error_handling.WriteJSON(w, http.StatusOK, room)
}

func (h *Handler) JoinRoom(w http.ResponseWriter, r *http.Request) {
	var req domain.JoinRoomRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		error_handling.WriteError(w, error_handling.BadRequestError("Invalid request body"))
		return
	}

	if appErr := h.Services.Room.JoinRoom(r.Context(), req.RoomID, req.PlayerID); appErr != nil {
		error_handling.WriteError(w, appErr)
		return
	}

	error_handling.WriteJSON(w, http.StatusOK, map[string]string{"message": "Player joined the room"})
}
