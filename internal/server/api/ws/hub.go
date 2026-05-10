package ws

import (
	"context"
	"encoding/json"
	"log"
	"sync"

	"game-server/internal/redis"
	"game-server/internal/server/domain"
)

type Hub struct {
	clients    map[*Client]bool
	rooms      map[string]map[*Client]bool
	register   chan *Client
	unregister chan *Client
	mu         sync.RWMutex
}

func NewHub() *Hub {
	return &Hub{
		clients:    make(map[*Client]bool),
		rooms:      make(map[string]map[*Client]bool),
		register:   make(chan *Client),
		unregister: make(chan *Client),
	}
}

func (h *Hub) Run(ctx context.Context, rdb *redis.Client) {
	go h.startRedisSubscriber(ctx, rdb)

	for {
		select {
		case client := <-h.register:
			h.mu.Lock()
			h.clients[client] = true
			if h.rooms[client.RoomID] == nil {
				h.rooms[client.RoomID] = make(map[*Client]bool)
			}
			h.rooms[client.RoomID][client] = true
			h.mu.Unlock()

		case client := <-h.unregister:
			h.mu.Lock()
			if _, ok := h.clients[client]; ok {
				delete(h.clients, client)
				if clients, ok := h.rooms[client.RoomID]; ok {
					delete(clients, client)
				}
				close(client.Send)
			}
			h.mu.Unlock()
		}
	}
}

func (h *Hub) BroadcastToRoom(roomID string, event domain.Event) {
	data, err := json.Marshal(event)
	if err != nil {
		return
	}

	h.mu.RLock()
	defer h.mu.RUnlock()

	if clients, ok := h.rooms[roomID]; ok {
		for client := range clients {
			select {
			case client.Send <- data:
			default:
				close(client.Send)
				delete(h.clients, client)
				delete(h.rooms[roomID], client)
			}
		}
	}
}

func (h *Hub) startRedisSubscriber(ctx context.Context, rdb *redis.Client) {
	if rdb == nil {
		return
	}

	pubsub := rdb.PSubscribe(ctx, "room:*")
	defer pubsub.Close()

	ch := pubsub.Channel()
	for msg := range ch {
		var event domain.Event
		if err := json.Unmarshal([]byte(msg.Payload), &event); err != nil {
			log.Printf("Failed to unmarshal Redis event: %v", err)
			continue
		}
		h.BroadcastToRoom(event.RoomID, event)
	}
}
