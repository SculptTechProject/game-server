package Types

type Room struct {
	ID        string   `json:"id"`
	Name      string   `json:"name"`
	PlayerIDs []string `json:"playerIds"`
}

type JoinRoomRequest struct {
	RoomID   string `json:"roomId"`
	PlayerID string `json:"playerId"`
}

type CreateRoomRequest struct {
	Name string
}
