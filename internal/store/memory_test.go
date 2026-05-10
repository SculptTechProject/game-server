package store

import (
	"context"
	"testing"

	"game-server/internal/server/domain"
)

func TestMemoryStore_CreateAndGetPlayer(t *testing.T) {
	s := NewMemoryStore()
	ctx := context.Background()

	player := domain.Player{ID: "p1", Nickname: "alice", HP: 100}
	if err := s.CreatePlayer(ctx, player); err != nil {
		t.Fatal(err)
	}

	got, err := s.GetPlayer(ctx, "p1")
	if err != nil {
		t.Fatal(err)
	}
	if got.Nickname != "alice" {
		t.Errorf("expected nickname 'alice', got '%s'", got.Nickname)
	}
	if got.HP != 100 {
		t.Errorf("expected HP 100, got %d", got.HP)
	}
}

func TestMemoryStore_GetPlayer_NotFound(t *testing.T) {
	s := NewMemoryStore()
	_, err := s.GetPlayer(context.Background(), "nonexistent")
	if err != ErrNotFound {
		t.Errorf("expected ErrNotFound, got %v", err)
	}
}

func TestMemoryStore_CreateAndGetRoom(t *testing.T) {
	s := NewMemoryStore()
	ctx := context.Background()

	room := domain.Room{ID: "r1", Name: "lobby", MaxPlayers: 10}
	if err := s.CreateRoom(ctx, room); err != nil {
		t.Fatal(err)
	}

	got, err := s.GetRoom(ctx, "r1")
	if err != nil {
		t.Fatal(err)
	}
	if got.Name != "lobby" {
		t.Errorf("expected name 'lobby', got '%s'", got.Name)
	}
	if got.MaxPlayers != 10 {
		t.Errorf("expected MaxPlayers 10, got %d", got.MaxPlayers)
	}
}

func TestMemoryStore_GetRoom_NotFound(t *testing.T) {
	s := NewMemoryStore()
	_, err := s.GetRoom(context.Background(), "nonexistent")
	if err != ErrNotFound {
		t.Errorf("expected ErrNotFound, got %v", err)
	}
}

func TestMemoryStore_AddPlayerToRoom(t *testing.T) {
	s := NewMemoryStore()
	ctx := context.Background()

	room := domain.Room{ID: "r1", Name: "lobby", MaxPlayers: 10}
	if err := s.CreateRoom(ctx, room); err != nil {
		t.Fatal(err)
	}

	if err := s.AddPlayerToRoom(ctx, "r1", "p1"); err != nil {
		t.Fatal(err)
	}
	if err := s.AddPlayerToRoom(ctx, "r1", "p2"); err != nil {
		t.Fatal(err)
	}

	got, err := s.GetRoom(ctx, "r1")
	if err != nil {
		t.Fatal(err)
	}
	if len(got.PlayerIDs) != 2 {
		t.Errorf("expected 2 players, got %d", len(got.PlayerIDs))
	}
}

func TestMemoryStore_AddPlayerToRoom_RoomNotFound(t *testing.T) {
	s := NewMemoryStore()
	err := s.AddPlayerToRoom(context.Background(), "nonexistent", "p1")
	if err != ErrNotFound {
		t.Errorf("expected ErrNotFound, got %v", err)
	}
}

func TestMemoryStore_ConcurrentAccess(t *testing.T) {
	s := NewMemoryStore()
	ctx := context.Background()

	room := domain.Room{ID: "r1", Name: "stress", MaxPlayers: 100}
	if err := s.CreateRoom(ctx, room); err != nil {
		t.Fatal(err)
	}

	errs := make(chan error, 100)
	for i := 0; i < 100; i++ {
		go func(id int) {
			errs <- s.AddPlayerToRoom(ctx, "r1", "p"+string(rune(id)))
		}(i)
	}

	for i := 0; i < 100; i++ {
		if err := <-errs; err != nil {
			t.Error(err)
		}
	}

	got, err := s.GetRoom(ctx, "r1")
	if err != nil {
		t.Fatal(err)
	}
	if len(got.PlayerIDs) != 100 {
		t.Errorf("expected 100 players, got %d", len(got.PlayerIDs))
	}
}
