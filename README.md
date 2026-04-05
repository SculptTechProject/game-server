

# game-server

Lightweight multiplayer game lobby server built with Go, Swagger, and planned WebSocket support.

## Overview

`game-server` is a backend project written in Go for managing game rooms and players.
It started as a simple HTTP API project and is being extended toward real-time multiplayer communication with WebSockets.

The main goal of this repository is to practice backend engineering through a practical project instead of building only basic CRUD apps.

## Features

### Current

- HTTP API built with Go
- room creation flow
- joining players to rooms
- room state fetching
- Swagger documentation
- UUID-based identifiers
- project structure split into `cmd`, `internal`, and `docs`

### Planned

- WebSocket support for real-time communication
- live room events
- player disconnect handling
- better validation and error responses
- tests
- persistence layer
- Docker support

## Tech stack

- Go 1.25
- `net/http`
- `github.com/google/uuid`
- `github.com/swaggo/swag`
- `github.com/swaggo/http-swagger`

## Project structure

```text
game-server/
├── cmd/
│   └── game-server/
│       └── main.go
├── internal/
│   └── ...
├── docs/
│   └── ...
├── go.mod
└── go.sum
```

## Getting started

### 1. Clone the repository

```bash
git clone https://github.com/SculptTechProject/game-server.git
cd game-server
```

### 2. Install dependencies

```bash
go mod tidy
```

### 3. Run the server

```bash
go run ./cmd/game-server
```

The server should start locally on:

```text
http://localhost:8080
```

## Swagger documentation

Swagger UI should be available at:

```text
http://localhost:8080/swagger/index.html
```

If you change endpoint annotations, regenerate the docs with:

```bash
swag init -g cmd/game-server/main.go
```

## API direction

This project is centered around a simple multiplayer room flow:

1. Create a room
2. Join a player to a room
3. Fetch room state
4. Extend communication toward real-time gameplay events

Example endpoint ideas used in the project:

- `POST /create-room`
- `POST /join-room`
- `GET /get-room`
- `GET /health`

## Real-time roadmap

The next major step for this project is WebSocket support.
The idea is to move from a simple request-response lobby API into a real-time multiplayer server foundation.

Planned WebSocket-related goals:

- keep live player connections per room
- broadcast join and leave events
- push room state updates instantly
- support game events without constant polling
- prepare the codebase for real-time match logic

## Why this project exists

I built this project to learn backend development in Go in a more practical way.
Instead of stopping at basic REST endpoints, I want to grow it into something closer to a real multiplayer backend.

This project helps me practice:

- backend architecture in Go
- API design
- shared state management
- Swagger integration
- real-time server thinking
- evolving a project from prototype to stronger backend foundation

## Roadmap

- [x] Basic HTTP server
- [x] Room creation
- [x] Join room flow
- [x] Swagger integration
- [ ] Better request validation
- [ ] Improved error handling
- [ ] Unit and integration tests
- [ ] WebSocket connections
- [ ] Real-time room events
- [ ] Persistent storage
- [ ] Docker setup

## Contributing

Suggestions, issues, and improvements are welcome.

## Author

Created by [SculptTechProject](https://github.com/SculptTechProject)