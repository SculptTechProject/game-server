package domain

import "time"

type EventType string

const (
	EventPlayerJoined  EventType = "player_joined"
	EventPlayerLeft    EventType = "player_left"
	EventPlayerMoved   EventType = "player_moved"
	EventRoomCreated   EventType = "room_created"
	EventRoomState     EventType = "room_state"
	EventCoinCollected EventType = "coin_collected"
	EventError         EventType = "error"
)

type Event struct {
	Type      EventType `json:"type"`
	RoomID    string    `json:"roomId"`
	PlayerID  string    `json:"playerId,omitempty"`
	Payload   any       `json:"payload,omitempty"`
	Timestamp time.Time `json:"timestamp"`
}
