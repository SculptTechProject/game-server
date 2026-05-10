package server

import (
	"log"
	"net/http"
)

func StartServer(handler http.Handler, port string) {
	addr := ":" + port
	log.Printf("Server started on %s", addr)
	log.Fatal(http.ListenAndServe(addr, handler))
}
