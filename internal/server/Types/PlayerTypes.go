package Types

type Player struct {
	ID       string `json:"id"`
	Nickname string `json:"nickname"`
	HP       int    `json:"hp"`
}

type CreatePlayerRequest struct {
	Nickname string `json:"nickname"`
}
