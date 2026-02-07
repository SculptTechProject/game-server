package server

type Room struct {
	ID        string   `json:"id"`
	Name      string   `json:"name"`
	PlayerIDs []string `json:"playerIds"`
}

type joinRoomRequest struct {
	RoomID   string `json:"roomId"`
	PlayerID string `json:"playerId"`
}

type createRoomRequest struct {
	Name string
}
