package http

import "net/http"

func RegisterRoutes() {
	//players
	http.HandleFunc("/create-player", CreatePlayer)
	http.HandleFunc("/get-player", GetPlayer)

	// rooms
	http.HandleFunc("/create-room", CreateRoom)
	http.HandleFunc("/get-room", GetRoom)
	http.HandleFunc("/join-room", JoinRoom)
}
