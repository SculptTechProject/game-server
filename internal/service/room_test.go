package service

import (
	"context"
	"fmt"
	"testing"

	"game-server/internal/error_handling"
	"game-server/internal/server/domain"
	"game-server/internal/store"
)

func setupRoomService(t *testing.T) (*RoomService, store.Store) {
	t.Helper()
	s := store.NewMemoryStore()
	svc := &RoomService{store: s}
	return svc, s
}

func TestRoomService_CreateRoom(t *testing.T) {
	svc, _ := setupRoomService(t)

	t.Run("success", func(t *testing.T) {
		room, appErr := svc.CreateRoom(context.Background(), "lobby", 10)
		if appErr != nil {
			t.Fatal(appErr)
		}
		if room.Name != "lobby" {
			t.Errorf("expected name 'lobby', got '%s'", room.Name)
		}
		if room.MaxPlayers != 10 {
			t.Errorf("expected MaxPlayers 10, got %d", room.MaxPlayers)
		}
		if room.ID == "" {
			t.Error("expected non-empty ID")
		}
	})

	t.Run("empty name", func(t *testing.T) {
		_, appErr := svc.CreateRoom(context.Background(), "", 5)
		if appErr == nil {
			t.Fatal("expected error")
		}
		if appErr.Code != error_handling.CodeBadRequest {
			t.Errorf("expected CodeBadRequest, got %s", appErr.Code)
		}
	})
}

func TestRoomService_GetRoom(t *testing.T) {
	svc, _ := setupRoomService(t)

	room, _ := svc.CreateRoom(context.Background(), "arena", 20)
	createdID := room.ID

	t.Run("found", func(t *testing.T) {
		got, appErr := svc.GetRoom(context.Background(), createdID)
		if appErr != nil {
			t.Fatal(appErr)
		}
		if got.Name != "arena" {
			t.Errorf("expected name 'arena', got '%s'", got.Name)
		}
	})

	t.Run("not found", func(t *testing.T) {
		_, appErr := svc.GetRoom(context.Background(), "nonexistent")
		if appErr == nil {
			t.Fatal("expected error")
		}
		if appErr.Code != error_handling.CodeNotFound {
			t.Errorf("expected CodeNotFound, got %s", appErr.Code)
		}
	})

	t.Run("empty id", func(t *testing.T) {
		_, appErr := svc.GetRoom(context.Background(), "")
		if appErr == nil {
			t.Fatal("expected error")
		}
		if appErr.Code != error_handling.CodeBadRequest {
			t.Errorf("expected CodeBadRequest, got %s", appErr.Code)
		}
	})
}

func TestRoomService_JoinRoom(t *testing.T) {
	svc, store := setupRoomService(t)

	room, _ := svc.CreateRoom(context.Background(), "battle", 2)

	t.Run("first player joins", func(t *testing.T) {
		appErr := svc.JoinRoom(context.Background(), room.ID, "p1")
		if appErr != nil {
			t.Fatal(appErr)
		}

		got, _ := store.GetRoom(context.Background(), room.ID)
		if len(got.PlayerIDs) != 1 {
			t.Errorf("expected 1 player, got %d", len(got.PlayerIDs))
		}
	})

	t.Run("second player joins", func(t *testing.T) {
		appErr := svc.JoinRoom(context.Background(), room.ID, "p2")
		if appErr != nil {
			t.Fatal(appErr)
		}
	})

	t.Run("duplicate player gets conflict", func(t *testing.T) {
		appErr := svc.JoinRoom(context.Background(), room.ID, "p1")
		if appErr == nil {
			t.Fatal("expected error for duplicate player")
		}
		if appErr.Code != error_handling.CodeConflict {
			t.Errorf("expected CodeConflict, got %s", appErr.Code)
		}
	})

	t.Run("room full", func(t *testing.T) {
		appErr := svc.JoinRoom(context.Background(), room.ID, "p3")
		if appErr == nil {
			t.Fatal("expected error for full room")
		}
		if appErr.Code != error_handling.CodeConflict {
			t.Errorf("expected CodeConflict, got %s", appErr.Code)
		}
	})

	t.Run("room not found", func(t *testing.T) {
		appErr := svc.JoinRoom(context.Background(), "nonexistent", "p1")
		if appErr == nil {
			t.Fatal("expected error")
		}
		if appErr.Code != error_handling.CodeNotFound {
			t.Errorf("expected CodeNotFound, got %s", appErr.Code)
		}
	})

	t.Run("empty room id", func(t *testing.T) {
		appErr := svc.JoinRoom(context.Background(), "", "p1")
		if appErr == nil {
			t.Fatal("expected error")
		}
		if appErr.Code != error_handling.CodeBadRequest {
			t.Errorf("expected CodeBadRequest, got %s", appErr.Code)
		}
	})

	t.Run("empty player id", func(t *testing.T) {
		appErr := svc.JoinRoom(context.Background(), room.ID, "")
		if appErr == nil {
			t.Fatal("expected error")
		}
		if appErr.Code != error_handling.CodeBadRequest {
			t.Errorf("expected CodeBadRequest, got %s", appErr.Code)
		}
	})
}

func TestRoomService_JoinRoom_EventPublished(t *testing.T) {
	s := store.NewMemoryStore()
	svc := &RoomService{store: s}

	room, _ := svc.CreateRoom(context.Background(), "events", 5)

	appErr := svc.JoinRoom(context.Background(), room.ID, "p1")
	if appErr != nil {
		t.Fatal(appErr)
	}

	got, _ := s.GetRoom(context.Background(), room.ID)
	if len(got.PlayerIDs) != 1 {
		t.Errorf("expected 1 player in room, got %d", len(got.PlayerIDs))
	}
}

func TestRoomService_UnlimitedCapacity(t *testing.T) {
	svc, s := setupRoomService(t)
	room, _ := svc.CreateRoom(context.Background(), "unlimited", 0)

	for i := 0; i < 50; i++ {
		pid := fmt.Sprintf("p%d", i)
		player := domain.Player{ID: pid, Nickname: pid}
		s.CreatePlayer(context.Background(), player)
	}

	for i := 0; i < 50; i++ {
		pid := fmt.Sprintf("p%d", i)
		if appErr := svc.JoinRoom(context.Background(), room.ID, pid); appErr != nil {
			t.Fatalf("player %d failed to join: %v", i, appErr)
		}
	}

	got, _ := s.GetRoom(context.Background(), room.ID)
	if len(got.PlayerIDs) != 50 {
		t.Errorf("expected 50 players, got %d", len(got.PlayerIDs))
	}
}
