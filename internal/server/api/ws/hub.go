package ws

import (
	"context"
	"encoding/json"
	"log"
	"sync"
	"time"

	"game-server/internal/redis"
	"game-server/internal/server/domain"
)

type PlayerPos struct {
	X float64 `json:"x"`
	Y float64 `json:"y"`
}

type Hub struct {
	clients         map[*Client]bool
	rooms           map[string]map[*Client]bool
	register        chan *Client
	unregister      chan *Client
	mu              sync.RWMutex
	playerPositions map[string]map[string]PlayerPos
	posMu           sync.RWMutex
	nicknames       map[string]map[string]string
	nickMu          sync.RWMutex
}

func NewHub() *Hub {
	return &Hub{
		clients:         make(map[*Client]bool),
		rooms:           make(map[string]map[*Client]bool),
		register:        make(chan *Client),
		unregister:      make(chan *Client),
		playerPositions: make(map[string]map[string]PlayerPos),
		nicknames:       make(map[string]map[string]string),
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

			nickname := client.Nickname
			if nickname == "" {
				nickname = client.PlayerID
			}
			h.nickMu.Lock()
			if h.nicknames[client.RoomID] == nil {
				h.nicknames[client.RoomID] = make(map[string]string)
			}
			h.nicknames[client.RoomID][client.PlayerID] = nickname
			h.nickMu.Unlock()

			h.sendRoomState(client)

			h.BroadcastToRoom(client.RoomID, domain.Event{
				Type:      domain.EventPlayerJoined,
				RoomID:    client.RoomID,
				PlayerID:  client.PlayerID,
				Payload:   map[string]string{"nickname": nickname},
				Timestamp: time.Now(),
			})

		case client := <-h.unregister:
			h.mu.Lock()
			if _, ok := h.clients[client]; ok {
				delete(h.clients, client)
				if clients, ok := h.rooms[client.RoomID]; ok {
					delete(clients, client)
					if len(clients) == 0 {
						h.posMu.Lock()
						delete(h.playerPositions, client.RoomID)
						h.posMu.Unlock()
						h.nickMu.Lock()
						delete(h.nicknames, client.RoomID)
						h.nickMu.Unlock()
					} else {
						h.nickMu.Lock()
						delete(h.nicknames[client.RoomID], client.PlayerID)
						h.nickMu.Unlock()
					}
				}
				close(client.Send)
			}
			h.mu.Unlock()

			h.BroadcastToRoom(client.RoomID, domain.Event{
				Type:      domain.EventPlayerLeft,
				RoomID:    client.RoomID,
				PlayerID:  client.PlayerID,
				Timestamp: time.Now(),
			})
		}
	}
}

func (h *Hub) sendRoomState(client *Client) {
	h.posMu.RLock()
	positions, ok := h.playerPositions[client.RoomID]
	h.posMu.RUnlock()

	if !ok || len(positions) == 0 {
		return
	}

	h.nickMu.RLock()
	nicks := h.nicknames[client.RoomID]
	h.nickMu.RUnlock()

	if len(nicks) <= 1 {
		return
	}

	payload := map[string]interface{}{
		"positions": positions,
		"nicknames": nicks,
	}

	event := domain.Event{
		Type:      domain.EventRoomState,
		RoomID:    client.RoomID,
		Payload:   payload,
		Timestamp: time.Now(),
	}

	data, err := json.Marshal(event)
	if err != nil {
		return
	}

	select {
	case client.Send <- data:
	default:
	}
}

func (h *Hub) HandleMessage(client *Client, data []byte) {
	var msg struct {
		Type   string   `json:"type"`
		X      *float64 `json:"x,omitempty"`
		Y      *float64 `json:"y,omitempty"`
		CoinId string   `json:"coinId,omitempty"`
	}
	if err := json.Unmarshal(data, &msg); err != nil {
		return
	}

	switch msg.Type {
	case "move":
		if msg.X == nil || msg.Y == nil {
			return
		}

		h.posMu.Lock()
		if h.playerPositions[client.RoomID] == nil {
			h.playerPositions[client.RoomID] = make(map[string]PlayerPos)
		}
		h.playerPositions[client.RoomID][client.PlayerID] = PlayerPos{X: *msg.X, Y: *msg.Y}
		h.posMu.Unlock()

		event := domain.Event{
			Type:     domain.EventPlayerMoved,
			RoomID:   client.RoomID,
			PlayerID: client.PlayerID,
			Payload: map[string]float64{
				"x": *msg.X,
				"y": *msg.Y,
			},
			Timestamp: time.Now(),
		}
		h.BroadcastToRoom(client.RoomID, event)

	case "collect":
		if msg.CoinId == "" {
			return
		}

		event := domain.Event{
			Type:     domain.EventCoinCollected,
			RoomID:   client.RoomID,
			PlayerID: client.PlayerID,
			Payload: map[string]interface{}{
				"coinId": msg.CoinId,
			},
			Timestamp: time.Now(),
		}
		h.BroadcastToRoom(client.RoomID, event)
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
