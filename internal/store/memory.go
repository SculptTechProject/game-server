package store

import (
	"context"
	"sync"

	"game-server/internal/server/domain"
)

type MemoryStore struct {
	mu      sync.RWMutex
	rooms   map[string]domain.Room
	players map[string]domain.Player
}

var Global = NewMemoryStore()

func NewMemoryStore() *MemoryStore {
	return &MemoryStore{
		rooms:   make(map[string]domain.Room),
		players: make(map[string]domain.Player),
	}
}

func (s *MemoryStore) CreatePlayer(_ context.Context, player domain.Player) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.players[player.ID] = player
	return nil
}

func (s *MemoryStore) GetPlayer(_ context.Context, id string) (domain.Player, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	player, ok := s.players[id]
	if !ok {
		return player, ErrNotFound
	}
	return player, nil
}

func (s *MemoryStore) CreateRoom(_ context.Context, room domain.Room) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.rooms[room.ID] = room
	return nil
}

func (s *MemoryStore) GetRoom(_ context.Context, id string) (domain.Room, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	room, ok := s.rooms[id]
	if !ok {
		return room, ErrNotFound
	}
	return room, nil
}

func (s *MemoryStore) AddPlayerToRoom(_ context.Context, roomID, playerID string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	room, ok := s.rooms[roomID]
	if !ok {
		return ErrNotFound
	}
	room.PlayerIDs = append(room.PlayerIDs, playerID)
	s.rooms[roomID] = room
	return nil
}
