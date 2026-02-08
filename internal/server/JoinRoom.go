package server

import (
	"encoding/json"
	"fmt"
	"game-server/internal/server/Types"
	"net/http"
)

// JoinRoom godoc
// @Summary Join room
// @Description Adds player to room if not already present
// @Tags rooms
// @Accept json
// @Produce plain
// @Param body body Types.JoinRoomRequest true "Join room request"
// @Success 200 {string} string
// @Failure 400 {string} string
// @Failure 404 {string} string
// @Router /join-room [post]
func JoinRoom() {
	http.HandleFunc("/join-room", func(w http.ResponseWriter, r *http.Request) {
		// check if method is post
		if r.Method != http.MethodPost {
			w.WriteHeader(http.StatusMethodNotAllowed)
			return
		}

		var req Types.JoinRoomRequest
		err := json.NewDecoder(r.Body).Decode(&req)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			_, _ = fmt.Fprintln(w, "Invalid request body")
			return
		}

		if req.RoomID == "" || req.PlayerID == "" {
			w.WriteHeader(http.StatusBadRequest)
			_, _ = fmt.Fprintln(w, "RoomID and PlayerID are required")
			return
		}

		Mu.Lock()
		room, ok := Rooms[req.RoomID]
		if !ok {
			Mu.Unlock()
			w.WriteHeader(http.StatusNotFound)
			_, _ = fmt.Fprintf(w, "Room not found: %s\n", req.RoomID)
			return
		}

		// add player to room if not already present
		alreadyInRoom := false
		for _, pid := range room.PlayerIDs {
			if pid == req.PlayerID {
				alreadyInRoom = true
				break
			}
		}
		if !alreadyInRoom {
			room.PlayerIDs = append(room.PlayerIDs, req.PlayerID)
		}

		Rooms[req.RoomID] = room
		Mu.Unlock()

		w.WriteHeader(http.StatusOK)
		_, _ = fmt.Fprintln(w, "Player joined the room")
	})
}
