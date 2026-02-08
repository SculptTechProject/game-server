package http

import (
	"encoding/json"
	"fmt"
	"game-server/internal/server/domain"
	id2 "game-server/internal/server/id"
	"game-server/internal/store"
	"net/http"
	"strings"
)

// CreatePlayer godoc
// @Summary Create a new player
// @Description Creates a player with a nickname and returns the created player (with generated id)
// @Tags players
// @Accept json
// @Produce json
// @Param body body domain.CreatePlayerRequest true "Create player request"
// @Success 201 {object} domain.Player
// @Failure 400 {string} string
// @Failure 500 {string} string
// @Router /create-player [post]
func CreatePlayer(w http.ResponseWriter, r *http.Request) {
	// check if method is post
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	var req domain.CreatePlayerRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		_, _ = fmt.Fprintln(w, "Invalid request body")
		return
	}

	req.Nickname = strings.TrimSpace(req.Nickname)
	if req.Nickname == "" {
		w.WriteHeader(http.StatusBadRequest)
		_, _ = fmt.Fprintln(w, "Nickname is required")
		return
	}

	id, err := id2.GenerateUniqeID()
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		_, _ = fmt.Fprintln(w, "Could not generate player id")
		return
	}

	player := domain.Player{ID: id, Nickname: req.Nickname, HP: 100}

	Mu.Lock()
	store.Players[id] = player
	Mu.Unlock()

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	_ = json.NewEncoder(w).Encode(player)
}

// GetPlayer godoc
// @Summary Get player
// @Description Returns player details by playerId
// @Tags players
// @Produce json
// @Param playerId query string true "Player ID"
// @Success 200 {object} domain.Player
// @Failure 400 {string} string
// @Failure 404 {string} string
// @Router /get-player [get]
func GetPlayer(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	playerID := strings.TrimSpace(r.URL.Query().Get("playerId"))
	if playerID == "" {
		w.WriteHeader(http.StatusBadRequest)
		_, _ = fmt.Fprintln(w, "playerId query param is required")
		return
	}

	Mu.RLock()
	player, ok := store.Players[playerID]
	Mu.RUnlock()

	if !ok {
		w.WriteHeader(http.StatusNotFound)
		_, _ = fmt.Fprintf(w, "Player not found: %s\n", playerID)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(player)
}
