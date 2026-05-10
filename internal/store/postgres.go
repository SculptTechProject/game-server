package store

import (
	"context"
	"errors"
	"fmt"

	"game-server/internal/database"
	"game-server/internal/server/domain"

	"github.com/jackc/pgx/v5"
)

type PostgresStore struct {
	db *database.Postgres
}

func NewPostgresStore(db *database.Postgres) *PostgresStore {
	return &PostgresStore{db: db}
}

func (s *PostgresStore) CreatePlayer(ctx context.Context, player domain.Player) error {
	_, err := s.db.Pool.Exec(ctx,
		`INSERT INTO players (id, nickname, hp) VALUES ($1, $2, $3)`,
		player.ID, player.Nickname, player.HP,
	)
	if err != nil {
		return fmt.Errorf("insert player: %w", err)
	}
	return nil
}

func (s *PostgresStore) GetPlayer(ctx context.Context, id string) (domain.Player, error) {
	var player domain.Player
	err := s.db.Pool.QueryRow(ctx,
		`SELECT id, nickname, hp FROM players WHERE id = $1`, id,
	).Scan(&player.ID, &player.Nickname, &player.HP)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return player, ErrNotFound
		}
		return player, fmt.Errorf("get player: %w", err)
	}
	return player, nil
}

func (s *PostgresStore) CreateRoom(ctx context.Context, room domain.Room) error {
	_, err := s.db.Pool.Exec(ctx,
		`INSERT INTO rooms (id, name, max_players) VALUES ($1, $2, $3)`,
		room.ID, room.Name, room.MaxPlayers,
	)
	if err != nil {
		return fmt.Errorf("insert room: %w", err)
	}
	return nil
}

func (s *PostgresStore) GetRoom(ctx context.Context, id string) (domain.Room, error) {
	var room domain.Room
	err := s.db.Pool.QueryRow(ctx,
		`SELECT id, name, max_players FROM rooms WHERE id = $1`, id,
	).Scan(&room.ID, &room.Name, &room.MaxPlayers)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return room, ErrNotFound
		}
		return room, fmt.Errorf("get room: %w", err)
	}

	rows, err := s.db.Pool.Query(ctx,
		`SELECT player_id FROM room_players WHERE room_id = $1 ORDER BY joined_at`, id,
	)
	if err != nil {
		return room, fmt.Errorf("get room players: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		var playerID string
		if err := rows.Scan(&playerID); err != nil {
			return room, fmt.Errorf("scan player id: %w", err)
		}
		room.PlayerIDs = append(room.PlayerIDs, playerID)
	}

	return room, rows.Err()
}

func (s *PostgresStore) AddPlayerToRoom(ctx context.Context, roomID, playerID string) error {
	_, err := s.db.Pool.Exec(ctx,
		`INSERT INTO room_players (room_id, player_id) VALUES ($1, $2)`,
		roomID, playerID,
	)
	if err != nil {
		return fmt.Errorf("add player to room: %w", err)
	}
	return nil
}
