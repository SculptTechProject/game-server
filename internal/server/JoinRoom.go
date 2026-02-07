package server

import (
	"encoding/json"
	"fmt"
	"net/http"
)

func JoinRoom() {
	http.HandleFunc("/join-room", func(w http.ResponseWriter, r *http.Request) {
		// check if method is post
		if r.Method != http.MethodPost {
			w.WriteHeader(http.StatusMethodNotAllowed)
			return
		}

		var req joinRoomRequest
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

		mu.Lock()
		room, ok := rooms[req.RoomID]
		if !ok {
			mu.Unlock()
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

		rooms[req.RoomID] = room
		mu.Unlock()

		w.WriteHeader(http.StatusOK)
		_, _ = fmt.Fprintln(w, "Player joined the room")
	})
}
