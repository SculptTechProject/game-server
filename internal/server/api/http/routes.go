package http

import (
	"net/http"

	"game-server/internal/server/api/ws"
	"game-server/internal/server/middleware"
	"game-server/internal/service"

	_ "game-server/docs"
	httpSwagger "github.com/swaggo/http-swagger"
)

type Handler struct {
	Services *service.Services
	Hub      *ws.Hub
}

func (h *Handler) SetupHandler() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("/create-player", h.CreatePlayer)
	mux.HandleFunc("/get-player", h.GetPlayer)
	mux.HandleFunc("/create-room", h.CreateRoom)
	mux.HandleFunc("/get-room", h.GetRoom)
	mux.HandleFunc("/join-room", h.JoinRoom)
	mux.HandleFunc("/ws", h.ServeWS)
	mux.Handle("/swagger/", httpSwagger.WrapHandler)

	return middleware.Recovery(middleware.Logging(mux))
}

func (h *Handler) ServeWS(w http.ResponseWriter, r *http.Request) {
	ws.ServeWS(h.Hub, w, r)
}
