package service

import (
	"game-server/internal/redis"
	"game-server/internal/store"
)

type Services struct {
	Player *PlayerService
	Room   *RoomService
}

func NewServices(s store.Store, r *redis.Client) *Services {
	return &Services{
		Player: &PlayerService{store: s},
		Room:   &RoomService{store: s, redis: r},
	}
}
