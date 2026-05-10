// @title Game Server API
// @version 1.0
// @description Game lobby server with PostgreSQL, Redis, and WebSocket support
// @host localhost:8080
// @BasePath /
package main

import (
	"context"
	"log"

	"game-server/internal/config"
	"game-server/internal/database"
	redis2 "game-server/internal/redis"
	"game-server/internal/server"
	http2 "game-server/internal/server/api/http"
	"game-server/internal/server/api/ws"
	"game-server/internal/service"
	"game-server/internal/store"
)

func main() {
	cfg := config.Load()
	ctx := context.Background()

	pg, err := database.NewPostgres(ctx, cfg.PostgresDSN)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer pg.Close()

	var s store.Store = store.NewPostgresStore(pg)

	rdb, err := redis2.NewClient(ctx, cfg.RedisAddr, cfg.RedisPass)
	if err != nil {
		log.Printf("Warning: Redis not available, real-time features disabled: %v", err)
		rdb = nil
	}

	svcs := service.NewServices(s, rdb)

	hub := ws.NewHub()
	go hub.Run(ctx, rdb)

	handler := (&http2.Handler{
		Services: svcs,
		Hub:      hub,
	}).SetupHandler()

	server.StartServer(handler, cfg.ServerPort)
}
