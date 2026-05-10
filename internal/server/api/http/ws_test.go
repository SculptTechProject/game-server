package http

import (
	"context"
	"encoding/json"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"game-server/internal/server/api/ws"
	"game-server/internal/server/domain"
	"game-server/internal/service"
	"game-server/internal/store"

	"github.com/gorilla/websocket"
)

func setupWSHandler(t *testing.T) (*Handler, *ws.Hub) {
	t.Helper()
	s := store.NewMemoryStore()
	svcs := service.NewServices(s, nil)
	hub := ws.NewHub()
	ctx, cancel := context.WithCancel(context.Background())
	t.Cleanup(cancel)
	go hub.Run(ctx, nil)
	return &Handler{Services: svcs, Hub: hub}, hub
}

func wsConnect(t *testing.T, server *httptest.Server, roomID, playerID string) *websocket.Conn {
	t.Helper()
	u := "ws" + strings.TrimPrefix(server.URL, "http") + "/ws?roomId=" + roomID + "&playerId=" + playerID
	conn, _, err := websocket.DefaultDialer.Dial(u, nil)
	if err != nil {
		t.Fatal(err)
	}
	return conn
}

func readEvent(t *testing.T, conn *websocket.Conn, timeout time.Duration) domain.Event {
	t.Helper()
	conn.SetReadDeadline(time.Now().Add(timeout))
	_, msg, err := conn.ReadMessage()
	if err != nil {
		t.Fatalf("read message: %v", err)
	}
	var ev domain.Event
	if err := json.Unmarshal(msg, &ev); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	return ev
}

func TestWebSocket_ConnectReceivesPlayerJoined(t *testing.T) {
	h, _ := setupWSHandler(t)
	server := httptest.NewServer(h.SetupHandler())
	defer server.Close()

	conn := wsConnect(t, server, "test-room", "p1")
	defer conn.Close()

	ev := readEvent(t, conn, 2*time.Second)
	if ev.Type != domain.EventPlayerJoined {
		t.Errorf("expected player_joined, got %s", ev.Type)
	}
	if ev.PlayerID != "p1" {
		t.Errorf("expected playerID p1, got %s", ev.PlayerID)
	}
}

func TestWebSocket_ReceivesBroadcastAfterJoin(t *testing.T) {
	h, hub := setupWSHandler(t)
	server := httptest.NewServer(h.SetupHandler())
	defer server.Close()

	conn := wsConnect(t, server, "test-room", "p1")
	defer conn.Close()

	// Consume the initial player_joined event
	readEvent(t, conn, 2*time.Second)

	// Now broadcast
	event := domain.Event{
		Type:      domain.EventPlayerMoved,
		RoomID:    "test-room",
		PlayerID:  "p1",
		Payload:   map[string]float64{"x": 100, "y": 200},
		Timestamp: time.Now(),
	}
	hub.BroadcastToRoom("test-room", event)

	ev := readEvent(t, conn, 2*time.Second)
	if ev.Type != domain.EventPlayerMoved {
		t.Errorf("expected player_moved, got %s", ev.Type)
	}
	if ev.PlayerID != "p1" {
		t.Errorf("expected playerID p1, got %s", ev.PlayerID)
	}
}

func TestWebSocket_DifferentRoomsIsolated(t *testing.T) {
	h, hub := setupWSHandler(t)
	server := httptest.NewServer(h.SetupHandler())
	defer server.Close()

	connA := wsConnect(t, server, "room-a", "p1")
	defer connA.Close()
	connB := wsConnect(t, server, "room-b", "p2")
	defer connB.Close()

	// Consume player_joined events
	readEvent(t, connA, 2*time.Second)
	readEvent(t, connB, 2*time.Second)

	hub.BroadcastToRoom("room-a", domain.Event{
		Type:      domain.EventPlayerMoved,
		RoomID:    "room-a",
		PlayerID:  "p1",
		Timestamp: time.Now(),
	})

	// connA should receive it
	ev := readEvent(t, connA, 1*time.Second)
	if ev.RoomID != "room-a" {
		t.Errorf("expected room-a, got %s", ev.RoomID)
	}

	// connB should NOT receive it
	connB.SetReadDeadline(time.Now().Add(200 * time.Millisecond))
	_, _, err := connB.ReadMessage()
	if err == nil {
		t.Fatal("room-b client should NOT receive message from room-a")
	}
}

func TestWebSocket_MultipleClientsSameRoom(t *testing.T) {
	h, hub := setupWSHandler(t)
	server := httptest.NewServer(h.SetupHandler())
	defer server.Close()

	conn1 := wsConnect(t, server, "common", "p1")
	defer conn1.Close()
	conn2 := wsConnect(t, server, "common", "p2")
	defer conn2.Close()

	// Consume player_joined events
	readEvent(t, conn1, 2*time.Second)
	readEvent(t, conn2, 2*time.Second)

	hub.BroadcastToRoom("common", domain.Event{
		Type:      domain.EventRoomCreated,
		RoomID:    "common",
		Timestamp: time.Now(),
	})

	// Both clients should receive it
	readEvent(t, conn1, 1*time.Second)
	readEvent(t, conn2, 1*time.Second)
}

func TestWebSocket_SendMoveAndReceiveBroadcast(t *testing.T) {
	h, _ := setupWSHandler(t)
	server := httptest.NewServer(h.SetupHandler())
	defer server.Close()

	// Two players in same room
	conn1 := wsConnect(t, server, "arena", "p1")
	defer conn1.Close()
	conn2 := wsConnect(t, server, "arena", "p2")
	defer conn2.Close()

	// Consume player_joined
	readEvent(t, conn1, 2*time.Second)
	readEvent(t, conn2, 2*time.Second)

	// Player 1 sends a move
	move := `{"type":"move","x":150,"y":250}`
	conn1.SetWriteDeadline(time.Now().Add(2 * time.Second))
	if err := conn1.WriteMessage(websocket.TextMessage, []byte(move)); err != nil {
		t.Fatal(err)
	}

	// Player 2 should receive the move broadcast
	ev := readEvent(t, conn2, 2*time.Second)
	if ev.Type != domain.EventPlayerMoved {
		t.Errorf("expected player_moved, got %s", ev.Type)
	}
	if ev.PlayerID != "p1" {
		t.Errorf("expected playerID p1, got %s", ev.PlayerID)
	}
	if ev.RoomID != "arena" {
		t.Errorf("expected roomID arena, got %s", ev.RoomID)
	}
}
