package server

import (
	"encoding/json"
	"fmt"
	"game-server/internal/server/Types"
	"net/http"
	"strings"
)

// CreatePlayer godoc
// @Summary Create a new player
// @Description Creates a player with a nickname and returns the created player (with generated id)
// @Tags players
// @Accept json
// @Produce json
// @Param body body Types.CreatePlayerRequest true "Create player request"
// @Success 201 {object} Types.Player
// @Failure 400 {string} string
// @Failure 500 {string} string
// @Router /create-player [post]
func CreatePlayer() {
	http.HandleFunc("/create-player", func(w http.ResponseWriter, r *http.Request) {
		// check if method is post
		if r.Method != http.MethodPost {
			w.WriteHeader(http.StatusMethodNotAllowed)
			return
		}

		var req Types.CreatePlayerRequest
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

		id, err := GenerateUniqeID()
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			_, _ = fmt.Fprintln(w, "Could not generate player id")
			return
		}

		player := Types.Player{ID: id, Nickname: req.Nickname}

		Mu.Lock()
		players[id] = player
		Mu.Unlock()

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		_ = json.NewEncoder(w).Encode(player)
	})
}
