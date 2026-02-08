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

// CreateRoom godoc
// @Summary Create a new room
// @Description Creates a new room and returns its ID
// @Tags rooms
// @Accept json
// @Produce json
// @Param body body domain.CreateRoomRequest true "Create room request"
// @Success 201 {object} domain.Room
// @Failure 400 {string} string
// @Failure 500 {string} string
// @Router /create-room [post]
func CreateRoom(w http.ResponseWriter, r *http.Request) {
	// check if method is post
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	var req domain.CreateRoomRequest
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

	id, err := id2.GenerateUniqeID()
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		_, _ = fmt.Fprintln(w, "Could not generate room id")
		return
	}

	room := domain.Room{ID: id, Name: req.Name, PlayerIDs: []string{}}

	// synchronized add to rooms map
	Mu.Lock()
	store.Rooms[id] = room
	Mu.Unlock()

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	_ = json.NewEncoder(w).Encode(room)
}

// GetRoom godoc
// @Summary Get room
// @Description Returns room details by roomId
// @Tags rooms
// @Produce json
// @Param roomId query string true "Room ID"
// @Success 200 {object} domain.Room
// @Failure 400 {string} string
// @Failure 404 {string} string
// @Router /get-room [get]
func GetRoom(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	roomID := strings.TrimSpace(r.URL.Query().Get("roomId"))
	if roomID == "" {
		w.WriteHeader(http.StatusBadRequest)
		_, _ = fmt.Fprintln(w, "roomId query param is required")
		return
	}

	Mu.RLock()
	room, ok := store.Rooms[roomID]
	Mu.RUnlock()

	if !ok {
		w.WriteHeader(http.StatusNotFound)
		_, _ = fmt.Fprintf(w, "Room not found: %s\n", roomID)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(room)
}

// JoinRoom godoc
// @Summary Join room
// @Description Adds player to room if not already present
// @Tags rooms
// @Accept json
// @Produce plain
// @Param body body domain.JoinRoomRequest true "Join room request"
// @Success 200 {string} string
// @Failure 400 {string} string
// @Failure 404 {string} string
// @Router /join-room [post]
func JoinRoom(w http.ResponseWriter, r *http.Request) {
	// check if method is post
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	var req domain.JoinRoomRequest
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
	room, ok := store.Rooms[req.RoomID]
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

	store.Rooms[req.RoomID] = room
	Mu.Unlock()

	w.WriteHeader(http.StatusOK)
	_, _ = fmt.Fprintln(w, "Player joined the room")
}
