package main

import "game-server/internal/server"

func main() {
	server.EndpointHandler() // handle endpoints and logic

	////////////////
	//start server//
	////////////////
	server.StartServer()
	////////////////
	//start server//
	////////////////

}
