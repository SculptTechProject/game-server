package server

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
)

func GetRoom() {
	http.HandleFunc("/get-room", func(w http.ResponseWriter, r *http.Request) {
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
		room, ok := Rooms[roomID]
		Mu.RUnlock()

		if !ok {
			w.WriteHeader(http.StatusNotFound)
			_, _ = fmt.Fprintf(w, "Room not found: %s\n", roomID)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_ = json.NewEncoder(w).Encode(room)
	})
}
