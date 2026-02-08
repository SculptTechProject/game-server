package server

import (
	"encoding/json"
	"fmt"
	"game-server/internal/server/Types"
	"net/http"
	"strings"
)

// CreateRoom godoc
// @Summary Create a new room
// @Description Creates a new room and returns its ID
// @Tags rooms
// @Accept json
// @Produce json
// @Param body body Types.CreateRoomRequest true "Create room request"
// @Success 201 {object} Types.Room
// @Failure 400 {string} string
// @Failure 500 {string} string
// @Router /create-room [post]
func CreateRoom() {
	http.HandleFunc("/create-room", func(w http.ResponseWriter, r *http.Request) {
		// check if method is post
		if r.Method != http.MethodPost {
			w.WriteHeader(http.StatusMethodNotAllowed)
			return
		}

		var req Types.CreateRoomRequest
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

		room := Types.Room{ID: id, Name: req.Name, PlayerIDs: []string{}}

		// synchronized add to rooms map
		Mu.Lock()
		Rooms[id] = room
		Mu.Unlock()

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		_ = json.NewEncoder(w).Encode(room)
	})
}
