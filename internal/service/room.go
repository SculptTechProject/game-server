package service

import (
	"context"
	"encoding/json"
	"errors"
	"strings"
	"time"

	"game-server/internal/error_handling"
	"game-server/internal/redis"
	"game-server/internal/server/domain"
	"game-server/internal/server/id"
	"game-server/internal/store"
)

type RoomService struct {
	store store.Store
	redis *redis.Client
}

func (svc *RoomService) CreateRoom(ctx context.Context, name string, maxPlayers int) (*domain.Room, *error_handling.AppError) {
	name = strings.TrimSpace(name)
	if name == "" {
		return nil, error_handling.BadRequestError("Room name is required")
	}

	var room domain.Room
	for i := 0; i < 20; i++ {
		code := id.GenerateRoomCode()

		// Check if code is already taken
		_, err := svc.store.GetRoom(ctx, code)
		if err == nil {
			continue
		}
		if !errors.Is(err, store.ErrNotFound) {
			return nil, error_handling.InternalErrorWrap("Could not check room code", err)
		}

		room = domain.Room{
			ID:         code,
			Name:       name,
			PlayerIDs:  []string{},
			MaxPlayers: maxPlayers,
		}

		if err := svc.store.CreateRoom(ctx, room); err != nil {
			if strings.Contains(err.Error(), "already exists") {
				continue
			}
			return nil, error_handling.InternalErrorWrap("Could not create room", err)
		}

		svc.publishEvent(ctx, domain.Event{
			Type:      domain.EventRoomCreated,
			RoomID:    code,
			Timestamp: time.Now(),
		})

		return &room, nil
	}

	return nil, error_handling.InternalError("Could not generate unique room code after 20 attempts")
}

func (svc *RoomService) GetRoom(ctx context.Context, roomID string) (*domain.Room, *error_handling.AppError) {
	roomID = strings.TrimSpace(roomID)
	if roomID == "" {
		return nil, error_handling.BadRequestError("roomId is required")
	}

	room, err := svc.store.GetRoom(ctx, roomID)
	if err != nil {
		if errors.Is(err, store.ErrNotFound) {
			return nil, error_handling.NotFoundError("Room not found: " + roomID)
		}
		return nil, error_handling.InternalErrorWrap("Could not retrieve room", err)
	}

	return &room, nil
}

func (svc *RoomService) JoinRoom(ctx context.Context, roomID, playerID string) *error_handling.AppError {
	roomID = strings.TrimSpace(roomID)
	playerID = strings.TrimSpace(playerID)
	if roomID == "" || playerID == "" {
		return error_handling.BadRequestError("RoomID and PlayerID are required")
	}

	room, err := svc.store.GetRoom(ctx, roomID)
	if err != nil {
		if errors.Is(err, store.ErrNotFound) {
			return error_handling.NotFoundError("Room not found: " + roomID)
		}
		return error_handling.InternalErrorWrap("Could not retrieve room", err)
	}

	for _, pid := range room.PlayerIDs {
		if pid == playerID {
			return error_handling.ConflictError("Player is already in the room")
		}
	}

	if room.MaxPlayers > 0 && len(room.PlayerIDs) >= room.MaxPlayers {
		return error_handling.ConflictError("Room is full")
	}

	if err := svc.store.AddPlayerToRoom(ctx, roomID, playerID); err != nil {
		return error_handling.InternalErrorWrap("Could not join room", err)
	}

	svc.publishEvent(ctx, domain.Event{
		Type:      domain.EventPlayerJoined,
		RoomID:    roomID,
		PlayerID:  playerID,
		Timestamp: time.Now(),
	})

	return nil
}

func (svc *RoomService) publishEvent(ctx context.Context, event domain.Event) {
	if svc.redis == nil {
		return
	}
	data, err := json.Marshal(event)
	if err != nil {
		return
	}
	svc.redis.Publish(ctx, "room:"+event.RoomID, string(data))
}
