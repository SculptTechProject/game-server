package domain

type Room struct {
	ID         string   `json:"id"`
	Name       string   `json:"name"`
	PlayerIDs  []string `json:"playerIds"`
	MaxPlayers int      `json:"max-players"`
}

type JoinRoomRequest struct {
	RoomID   string `json:"roomId"`
	PlayerID string `json:"playerId"`
}

type CreateRoomRequest struct {
	Name       string `json:"name"`
	MaxPlayers int    `json:"max-players"`
}

type GetRoomRequest struct {
	RoomID string `json:"roomId"`
}
