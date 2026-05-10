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

func TestWebSocket_ConnectAndDisconnect(t *testing.T) {
	h, _ := setupWSHandler(t)
	server := httptest.NewServer(h.SetupHandler())
	defer server.Close()

	wsURL := "ws" + strings.TrimPrefix(server.URL, "http") + "/ws?roomId=test-room"

	conn, _, err := websocket.DefaultDialer.Dial(wsURL, nil)
	if err != nil {
		t.Fatal(err)
	}
	defer conn.Close()
}

func TestWebSocket_ReceivesRoomBroadcast(t *testing.T) {
	h, hub := setupWSHandler(t)
	server := httptest.NewServer(h.SetupHandler())
	defer server.Close()

	wsURL := "ws" + strings.TrimPrefix(server.URL, "http") + "/ws?roomId=test-room"

	conn, _, err := websocket.DefaultDialer.Dial(wsURL, nil)
	if err != nil {
		t.Fatal(err)
	}
	defer conn.Close()

	// Give hub time to register
	time.Sleep(50 * time.Millisecond)

	event := domain.Event{
		Type:      domain.EventPlayerJoined,
		RoomID:    "test-room",
		PlayerID:  "p1",
		Timestamp: time.Now(),
	}

	hub.BroadcastToRoom("test-room", event)

	conn.SetReadDeadline(time.Now().Add(2 * time.Second))
	_, msg, err := conn.ReadMessage()
	if err != nil {
		t.Fatal(err)
	}

	var received domain.Event
	if err := json.Unmarshal(msg, &received); err != nil {
		t.Fatal(err)
	}

	if received.Type != domain.EventPlayerJoined {
		t.Errorf("expected player_joined, got %s", received.Type)
	}
	if received.PlayerID != "p1" {
		t.Errorf("expected playerID p1, got %s", received.PlayerID)
	}
}

func TestWebSocket_DifferentRoomsIsolated(t *testing.T) {
	h, hub := setupWSHandler(t)
	server := httptest.NewServer(h.SetupHandler())
	defer server.Close()

	baseURL := "ws" + strings.TrimPrefix(server.URL, "http")

	connA, _, err := websocket.DefaultDialer.Dial(baseURL+"/ws?roomId=room-a", nil)
	if err != nil {
		t.Fatal(err)
	}
	defer connA.Close()

	connB, _, err := websocket.DefaultDialer.Dial(baseURL+"/ws?roomId=room-b", nil)
	if err != nil {
		t.Fatal(err)
	}
	defer connB.Close()

	time.Sleep(50 * time.Millisecond)

	hub.BroadcastToRoom("room-a", domain.Event{
		Type:      domain.EventPlayerJoined,
		RoomID:    "room-a",
		PlayerID:  "p1",
		Timestamp: time.Now(),
	})

	connA.SetReadDeadline(time.Now().Add(1 * time.Second))
	_, _, err = connA.ReadMessage()
	if err != nil {
		t.Fatal("room-a client should receive message:", err)
	}

	connB.SetReadDeadline(time.Now().Add(200 * time.Millisecond))
	_, _, err = connB.ReadMessage()
	if err == nil {
		t.Fatal("room-b client should NOT receive message from room-a")
	}
}

func TestWebSocket_MultipleClientsSameRoom(t *testing.T) {
	h, hub := setupWSHandler(t)
	server := httptest.NewServer(h.SetupHandler())
	defer server.Close()

	wsURL := "ws" + strings.TrimPrefix(server.URL, "http") + "/ws?roomId=common"

	conn1, _, err := websocket.DefaultDialer.Dial(wsURL, nil)
	if err != nil {
		t.Fatal(err)
	}
	defer conn1.Close()

	conn2, _, err := websocket.DefaultDialer.Dial(wsURL, nil)
	if err != nil {
		t.Fatal(err)
	}
	defer conn2.Close()

	time.Sleep(50 * time.Millisecond)

	hub.BroadcastToRoom("common", domain.Event{
		Type:      domain.EventRoomCreated,
		RoomID:    "common",
		Timestamp: time.Now(),
	})

	for i, conn := range []*websocket.Conn{conn1, conn2} {
		conn.SetReadDeadline(time.Now().Add(1 * time.Second))
		_, _, err := conn.ReadMessage()
		if err != nil {
			t.Fatalf("client %d should receive message: %v", i+1, err)
		}
	}
}
