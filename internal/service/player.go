package service

import (
	"context"
	"errors"

	"game-server/internal/error_handling"
	"game-server/internal/server/domain"
	"game-server/internal/server/id"
	"game-server/internal/store"
	"strings"
)

type PlayerService struct {
	store store.Store
}

func (svc *PlayerService) CreatePlayer(ctx context.Context, nickname string) (*domain.Player, *error_handling.AppError) {
	nickname = strings.TrimSpace(nickname)
	if nickname == "" {
		return nil, error_handling.BadRequestError("Nickname is required")
	}

	playerID, err := id.GenerateUniqueID()
	if err != nil {
		return nil, error_handling.InternalErrorWrap("Could not generate player ID", err)
	}

	player := domain.Player{
		ID:       playerID,
		Nickname: nickname,
		HP:       100,
	}

	if err := svc.store.CreatePlayer(ctx, player); err != nil {
		return nil, error_handling.InternalErrorWrap("Could not create player", err)
	}

	return &player, nil
}

func (svc *PlayerService) GetPlayer(ctx context.Context, playerID string) (*domain.Player, *error_handling.AppError) {
	playerID = strings.TrimSpace(playerID)
	if playerID == "" {
		return nil, error_handling.BadRequestError("playerId is required")
	}

	player, err := svc.store.GetPlayer(ctx, playerID)
	if err != nil {
		if errors.Is(err, store.ErrNotFound) {
			return nil, error_handling.NotFoundError("Player not found: " + playerID)
		}
		return nil, error_handling.InternalErrorWrap("Could not retrieve player", err)
	}

	return &player, nil
}
