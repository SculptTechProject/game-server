// @title Game Server API
// @version 1.0
// @description Simple game lobby server (rooms & players)
// @host localhost:8080
// @BasePath /
package main

import (
	"game-server/internal/server"
	http2 "game-server/internal/server/api/http"
	"net/http"

	_ "game-server/docs"

	httpSwagger "github.com/swaggo/http-swagger"
)

func main() {
	http2.RegisterRoutes() // handle endpoints and logic

	http.HandleFunc("/swagger/", httpSwagger.WrapHandler)

	////////////////
	//start server//
	////////////////
	server.StartServer()
	////////////////
	//start server//
	////////////////

}
