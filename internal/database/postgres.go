package database

import (
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

type Postgres struct {
	Pool *pgxpool.Pool
}

func NewPostgres(ctx context.Context, dsn string) (*Postgres, error) {
	cfg, err := pgxpool.ParseConfig(dsn)
	if err != nil {
		return nil, fmt.Errorf("parse config: %w", err)
	}
	cfg.MaxConns = 25
	cfg.MinConns = 5
	cfg.MaxConnLifetime = 30 * time.Minute
	cfg.MaxConnIdleTime = 5 * time.Minute

	pool, err := pgxpool.NewWithConfig(ctx, cfg)
	if err != nil {
		return nil, fmt.Errorf("create pool: %w", err)
	}

	if err := pool.Ping(ctx); err != nil {
		pool.Close()
		return nil, fmt.Errorf("ping: %w", err)
	}

	p := &Postgres{Pool: pool}
	if err := p.migrate(ctx); err != nil {
		pool.Close()
		return nil, fmt.Errorf("migrate: %w", err)
	}
	return p, nil
}

func (p *Postgres) Close() {
	p.Pool.Close()
}

func (p *Postgres) migrate(ctx context.Context) error {
	schema := `
	CREATE TABLE IF NOT EXISTS players (
		id TEXT PRIMARY KEY,
		nickname TEXT NOT NULL,
		hp INTEGER NOT NULL DEFAULT 100,
		created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
	);
	CREATE TABLE IF NOT EXISTS rooms (
		id TEXT PRIMARY KEY,
		name TEXT NOT NULL,
		max_players INTEGER NOT NULL DEFAULT 0,
		created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
	);
	CREATE TABLE IF NOT EXISTS room_players (
		room_id TEXT REFERENCES rooms(id) ON DELETE CASCADE,
		player_id TEXT REFERENCES players(id) ON DELETE CASCADE,
		joined_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
		PRIMARY KEY (room_id, player_id)
	);
	`
	_, err := p.Pool.Exec(ctx, schema)
	return err
}
