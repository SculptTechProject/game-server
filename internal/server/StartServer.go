package server

import (
	"fmt"
	"log"
	"net/http"
)

func StartServer() {
	http.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		_, err := fmt.Fprintf(w, "Server is working\n %d", http.StatusOK)
		if err != nil {
			return
		}
	})

	log.Println("Server started on localhost:8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}
