package server

type Player struct {
	ID       string
	Nickname string
	HP       int
}

type CreatePlayerRequest struct {
	Nickname string
}
