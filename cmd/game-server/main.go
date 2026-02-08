// @title Game Server API
// @version 1.0
// @description Simple game lobby server (rooms & players)
// @host localhost:8080
// @BasePath /
package main

import (
	"game-server/internal/server"
	"net/http"

	_ "game-server/docs"

	httpSwagger "github.com/swaggo/http-swagger"
)

func main() {
	server.EndpointHandler() // handle endpoints and logic

	http.HandleFunc("/swagger/", httpSwagger.WrapHandler)

	////////////////
	//start server//
	////////////////
	server.StartServer()
	////////////////
	//start server//
	////////////////

}
