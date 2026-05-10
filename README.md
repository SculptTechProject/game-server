# game-server

Multiplayer game lobby server built with Go. Supports HTTP API, WebSocket real-time events, PostgreSQL persistence, and Redis pub/sub for horizontal scaling.

## Architecture

```
┌─────────────┐     ┌──────────────┐     ┌────────────┐
│  Client     │────▶│  HTTP/WS     │────▶│  Service   │
│  (HTTP/WS)  │◀────│  (net/http)  │◀────│  Layer     │
└─────────────┘     └──────┬───────┘     └─────┬──────┘
                           │                    │
                           │              ┌─────▼──────┐
                           │              │   Store    │
                           │              │ (PostgreSQL│
                           │              │  /Memory)  │
                           │              └─────┬──────┘
                           │                    │
                    ┌──────▼───────┐            │
                    │  WebSocket   │            │
                    │    Hub       │            │
                    │ (Redis Pub/  │            │
                    │   Sub)       │            │
                    └──────────────┘            │
                                                │
                    ┌──────────────┐            │
                    │    Redis     │◀───────────┘
                    │  (Pub/Sub)   │
                    └──────────────┘
```

### Layer structure

| Layer | Path | Responsibility |
|---|---|---|
| **Entry point** | `cmd/game-server/main.go` | Config loading, DI wiring, server start |
| **Config** | `internal/config/` | Env-based configuration |
| **HTTP API** | `internal/server/api/http/` | Route handlers, request/response serialization |
| **WebSocket** | `internal/server/api/ws/` | Real-time connections, room broadcast |
| **Middleware** | `internal/server/middleware/` | Panic recovery, request logging |
| **Service** | `internal/service/` | Business logic, validation, event publishing |
| **Store** | `internal/store/` | Storage interface + PostgreSQL / in-memory implementations |
| **Domain** | `internal/server/domain/` | Shared types (Player, Room, Event) |
| **Database** | `internal/database/` | PostgreSQL connection pool + auto-migration |
| **Redis** | `internal/redis/` | Redis client wrapper for pub/sub |
| **Error handling** | `internal/error_handling/` | Standardized errors and JSON responses |

## Tech stack

- **Go 1.25** with `net/http`
- **PostgreSQL 16** via `jackc/pgx/v5` (connection pooling, prepared statements)
- **Redis 7** via `redis/go-redis/v9` (pub/sub for real-time events)
- **WebSocket** via `gorilla/websocket`
- **Swagger** via `swaggo/http-swagger`
- **UUID** via `google/uuid`

## Quick start

### Local development (no Docker)

Requires PostgreSQL and Redis running locally.

```bash
# Terminal 1: start dependencies
docker compose up postgres redis -d

# Terminal 2: start app
go run ./cmd/game-server
```

### Docker (everything)

```bash
docker compose up --build
```

The app is available at `http://localhost:8080`.

## Configuration

All via environment variables:

| Variable | Default | Description |
|---|---|---|
| `SERVER_PORT` | `8080` | HTTP server port |
| `POSTGRES_DSN` | `postgres://postgres:postgres@localhost:5432/game_server?sslmode=disable` | PostgreSQL connection string |
| `REDIS_ADDR` | `localhost:6379` | Redis address |
| `REDIS_PASS` | `""` | Redis password |

If Redis is unavailable, the app starts with a warning and real-time features are disabled.

## API

### Standard response format

All endpoints return JSON:

**Success:**
```json
{
  "success": true,
  "data": { ... }
}
```

**Error:**
```json
{
  "success": false,
  "error": {
    "code": "NOT_FOUND",
    "message": "Player not found: abc-123"
  }
}
```

### Error codes

| Code | HTTP Status | Description |
|---|---|---|
| `NOT_FOUND` | 404 | Resource not found |
| `BAD_REQUEST` | 400 | Invalid input |
| `CONFLICT` | 409 | Duplicate or capacity exceeded |
| `INTERNAL_ERROR` | 500 | Server-side failure |
| `METHOD_NOT_ALLOWED` | 405 | Wrong HTTP method |

### Endpoints

#### `POST /create-player`

Create a new player.

```bash
curl -X POST http://localhost:8080/create-player \
  -H "Content-Type: application/json" \
  -d '{"nickname":"alice"}'
```

Response: `201 Created`
```json
{
  "success": true,
  "data": {
    "id": "550e8400-e29b-41d4-a716-446655440000",
    "nickname": "alice",
    "hp": 100
  }
}
```

---

#### `GET /get-player?playerId={id}`

Get player details.

```bash
curl "http://localhost:8080/get-player?playerId=550e8400-e29b-41d4-a716-446655440000"
```

Response: `200 OK`
```json
{
  "success": true,
  "data": {
    "id": "550e8400-e29b-41d4-a716-446655440000",
    "nickname": "alice",
    "hp": 100
  }
}
```

---

#### `POST /create-room`

Create a new room.

```bash
curl -X POST http://localhost:8080/create-room \
  -H "Content-Type: application/json" \
  -d '{"name":"lobby","max-players":10}'
```

Response: `201 Created`
```json
{
  "success": true,
  "data": {
    "id": "660e8400-e29b-41d4-a716-446655440001",
    "name": "lobby",
    "playerIds": [],
    "max-players": 10
  }
}
```

---

#### `GET /get-room?roomId={id}`

Get room state including players.

```bash
curl "http://localhost:8080/get-room?roomId=660e8400-e29b-41d4-a716-446655440001"
```

Response: `200 OK`
```json
{
  "success": true,
  "data": {
    "id": "660e8400-e29b-41d4-a716-446655440001",
    "name": "lobby",
    "playerIds": ["550e8400-e29b-41d4-a716-446655440000"],
    "max-players": 10
  }
}
```

---

#### `POST /join-room`

Join a player to a room.

```bash
curl -X POST http://localhost:8080/join-room \
  -H "Content-Type: application/json" \
  -d '{"roomId":"660e8400-...","playerId":"550e8400-..."}'
```

Response: `200 OK`
```json
{
  "success": true,
  "data": {
    "message": "Player joined the room"
  }
}
```

Error cases: `400` (missing fields), `404` (room not found), `409` (already in room / room full).

---

## WebSocket

### Connect

```js
const ws = new WebSocket("ws://localhost:8080/ws?roomId=660e8400-...");

ws.onmessage = (event) => {
  const data = JSON.parse(event.data);
  console.log(data.type, data);
};
```

### Events

The server pushes events to all WebSocket clients connected to a room:

| Event type | Direction | Description |
|---|---|---|
| `room_created` | Server → Client | A new room was created |
| `player_joined` | Server → Client | A player joined, payload includes `nickname` |
| `player_left` | Server → Client | A player disconnected |
| `player_moved` | Server → Client | A player moved, payload `{x, y}` |
| `coin_collected` | Server → Client | A coin was collected, payload `{coinId}` |
| `room_state` | Server → Client | Current positions of all players (sent on join) |

**Client → Server messages:**

| Message type | Description |
|---|---|
| `{"type":"move","x":100,"y":200}` | Update player position |
| `{"type":"collect","coinId":"c01"}` | Collect a coin |

**Events are published via Redis pub/sub**, so the system scales horizontally.

---

## Game client

A web-based 2D multiplayer game lives in `client/`.

### Tech

- **Vue 3** + **TypeScript** + **Vite**
- Canvas 2D rendering
- WebSocket real-time communication

### Running

```bash
# Terminal 1 — start the Go server
go run ./cmd/game-server

# Terminal 2 — start the Vue dev server
cd client && npm run dev
```

Then open **http://localhost:5173** in two browser tabs.

### How to play

1. Enter a **nickname**
2. Leave **Room ID** empty to create a new room, or enter an existing room ID to join a friend
3. Click **Play**
4. Move your circle with **WASD** or **arrow keys**
5. Collect golden **★ coins** to earn points
6. Watch other players move in real-time

### Game features

| Feature | Description |
|---|---|
| **Real-time movement** | Position synced via WebSocket, ~20 updates/sec |
| **Coins** | 30 coins spawn across the map, respawn after 5s |
| **Score** | +10 points per coin, shown in HUD |
| **Minimap** | Top-right corner shows full map with player dots |
| **Particles** | Dust particles when walking, burst on coin collect |
| **Player colors** | Each player gets a deterministic color from their ID |
| **Nicknames** | Shown above each player circle |
| **Camera** | Follows your player smoothly |
| **Walls** | Brown border walls prevent leaving the map |
| **Trees** | Decorative obstacles at the corners |
| **Eyes** | Your character has cute eyes 👀 |

### Game structure

```
client/
├── index.html
├── package.json
├── vite.config.ts
├── tsconfig.json
├── src/
│   ├── main.ts
│   ├── App.vue              ← screen router (lobby / game)
│   ├── style.css
│   ├── components/
│   │   ├── Lobby.vue        ← nickname & room input
│   │   └── GameCanvas.vue   ← canvas rendering & game loop
│   └── game/
│       ├── types.ts         ← TypeScript interfaces
│       └── websocket.ts     ← WebSocket client
```

### World

- **Map size:** 2000 × 2000 px
- **Camera:** follows your player (viewport fills the window)
- **Boundaries:** brown walls, 24px thick
- **30 coins** at fixed positions, golden with glow effect
- **8 decorative trees** at map corners and edges

---

## Swagger

Swagger UI is available at:
```
http://localhost:8080/swagger/index.html
```

To regenerate docs:
```bash
swag init -g cmd/game-server/main.go
```

## Testing

```bash
# All tests
go test ./...

# With race detection
go test -race ./...

# Benchmarks
go test -bench=. ./...

# Verbose
go test -v ./...
```

### Test structure

| Test file | Type | What it covers |
|---|---|---|
| `internal/error_handling/errors_test.go` | Unit | Error constructors, unwrapping, HTTP status codes |
| `internal/store/memory_test.go` | Unit | Store CRUD, concurrent access safety |
| `internal/service/player_test.go` | Service | Player creation, retrieval, validation |
| `internal/service/room_test.go` | Service | Room CRUD, join logic, capacity, duplicates |
| `internal/server/api/http/handler_test.go` | Integration | Full HTTP call chain via `httptest` |
| `internal/server/api/http/ws_test.go` | Integration | WebSocket connect, room isolation, broadcast |
| `internal/server/api/http/benchmark_test.go` | Benchmark/Stress | Per-endpoint throughput, concurrent requests |

## Project structure

```
cmd/game-server/main.go
internal/
├── config/config.go
├── database/postgres.go
├── error_handling/
│   ├── codes.go
│   ├── errors.go
│   └── response.go
├── redis/client.go
├── server/
│   ├── StartServer.go
│   ├── api/
│   │   ├── http/
│   │   │   ├── routes.go
│   │   │   ├── players.go
│   │   │   ├── rooms.go
│   │   │   ├── handler_test.go
│   │   │   ├── ws_test.go
│   │   │   └── benchmark_test.go
│   │   └── ws/
│   │       ├── hub.go
│   │       ├── client.go
│   │       └── handler.go
│   ├── domain/
│   │   ├── PlayerTypes.go
│   │   ├── RoomTypes.go
│   │   └── EventTypes.go
│   ├── id/GenerateUniqueID.go
│   └── middleware/middleware.go
├── service/
│   ├── service.go
│   ├── player.go
│   ├── player_test.go
│   ├── room.go
│   └── room_test.go
└── store/
    ├── interface.go
    ├── memory.go
    ├── memory_test.go
    └── postgres.go
client/
├── package.json
├── vite.config.ts
├── index.html
└── src/
    ├── main.ts
    ├── App.vue
    ├── style.css
    ├── components/
    │   ├── Lobby.vue
    │   └── GameCanvas.vue
    └── game/
        ├── types.ts
        └── websocket.ts
Dockerfile
docker-compose.yml
```
