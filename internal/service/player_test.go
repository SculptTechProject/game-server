package service

import (
	"context"
	"testing"

	"game-server/internal/error_handling"
	"game-server/internal/store"
)

func TestPlayerService_CreatePlayer(t *testing.T) {
	s := store.NewMemoryStore()
	svc := &PlayerService{store: s}

	t.Run("success", func(t *testing.T) {
		player, appErr := svc.CreatePlayer(context.Background(), "alice")
		if appErr != nil {
			t.Fatal(appErr)
		}
		if player.Nickname != "alice" {
			t.Errorf("expected nickname 'alice', got '%s'", player.Nickname)
		}
		if player.HP != 100 {
			t.Errorf("expected HP 100, got %d", player.HP)
		}
		if player.ID == "" {
			t.Error("expected non-empty ID")
		}
	})

	t.Run("empty nickname", func(t *testing.T) {
		_, appErr := svc.CreatePlayer(context.Background(), "")
		if appErr == nil {
			t.Fatal("expected error")
		}
		if appErr.Code != error_handling.CodeBadRequest {
			t.Errorf("expected CodeBadRequest, got %s", appErr.Code)
		}
	})

	t.Run("whitespace nickname", func(t *testing.T) {
		_, appErr := svc.CreatePlayer(context.Background(), "   ")
		if appErr == nil {
			t.Fatal("expected error")
		}
	})
}

func TestPlayerService_GetPlayer(t *testing.T) {
	s := store.NewMemoryStore()
	svc := &PlayerService{store: s}

	player, _ := svc.CreatePlayer(context.Background(), "bob")
	createdID := player.ID

	t.Run("found", func(t *testing.T) {
		got, appErr := svc.GetPlayer(context.Background(), createdID)
		if appErr != nil {
			t.Fatal(appErr)
		}
		if got.Nickname != "bob" {
			t.Errorf("expected nickname 'bob', got '%s'", got.Nickname)
		}
	})

	t.Run("not found", func(t *testing.T) {
		_, appErr := svc.GetPlayer(context.Background(), "nonexistent")
		if appErr == nil {
			t.Fatal("expected error")
		}
		if appErr.Code != error_handling.CodeNotFound {
			t.Errorf("expected CodeNotFound, got %s", appErr.Code)
		}
	})

	t.Run("empty id", func(t *testing.T) {
		_, appErr := svc.GetPlayer(context.Background(), "")
		if appErr == nil {
			t.Fatal("expected error")
		}
		if appErr.Code != error_handling.CodeBadRequest {
			t.Errorf("expected CodeBadRequest, got %s", appErr.Code)
		}
	})
}
