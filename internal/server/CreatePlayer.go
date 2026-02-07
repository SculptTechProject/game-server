package server

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
)

func CreatePlayer() {
	http.HandleFunc("/create-player", func(w http.ResponseWriter, r *http.Request) {
		// check if method is post
		if r.Method != http.MethodPost {
			w.WriteHeader(http.StatusMethodNotAllowed)
			return
		}

		var req CreatePlayerRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			w.WriteHeader(http.StatusBadRequest)
			_, _ = fmt.Fprintln(w, "Invalid request body")
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
		}

		player := Player{ID: id, Nickname: req.Nickname}

		mu.Lock()
		players[id] = player
		mu.Unlock()

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		_ = json.NewEncoder(w).Encode(player)
	})
}
