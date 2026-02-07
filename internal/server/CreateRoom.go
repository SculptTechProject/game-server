package server

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
)

func CreateRoom() {
	http.HandleFunc("/create-room", func(w http.ResponseWriter, r *http.Request) {
		// check if method is post
		if r.Method != http.MethodPost {
			w.WriteHeader(http.StatusMethodNotAllowed)
			return
		}

		var req createRoomRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			w.WriteHeader(http.StatusBadRequest)
			_, _ = fmt.Fprintln(w, "Invalid request body")
			return
		}

		req.Name = strings.TrimSpace(req.Name)
		if req.Name == "" {
			w.WriteHeader(http.StatusBadRequest)
			_, _ = fmt.Fprintln(w, "Room name is required")
			return
		}

		id, err := GenerateUniqeID()
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			_, _ = fmt.Fprintln(w, "Could not generate room id")
			return
		}

		room := Room{ID: id, Name: req.Name, PlayerIDs: []string{}}

		// synchronized add to rooms map
		mu.Lock()
		rooms[id] = room
		mu.Unlock()

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		_ = json.NewEncoder(w).Encode(room)
	})
}
