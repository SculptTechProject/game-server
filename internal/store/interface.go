package store

import (
	"context"
	"errors"

	"game-server/internal/server/domain"
)

var ErrNotFound = errors.New("not found")

type Store interface {
	CreatePlayer(ctx context.Context, player domain.Player) error
	GetPlayer(ctx context.Context, id string) (domain.Player, error)
	CreateRoom(ctx context.Context, room domain.Room) error
	GetRoom(ctx context.Context, id string) (domain.Room, error)
	AddPlayerToRoom(ctx context.Context, roomID, playerID string) error
}
