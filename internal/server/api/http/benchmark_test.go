package http

import (
	"bytes"
	"fmt"
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"

	"game-server/internal/service"
	"game-server/internal/store"
)

func BenchmarkCreatePlayer(b *testing.B) {
	s := store.NewMemoryStore()
	svcs := service.NewServices(s, nil)
	h := &Handler{Services: svcs}
	server := httptest.NewServer(h.SetupHandler())
	defer server.Close()

	body := []byte(`{"nickname":"benchmark"}`)
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		resp, err := http.Post(server.URL+"/create-player", "application/json", bytes.NewReader(body))
		if err != nil {
			b.Fatal(err)
		}
		resp.Body.Close()
		if resp.StatusCode != http.StatusCreated {
			b.Fatalf("expected 201, got %d", resp.StatusCode)
		}
	}
}

func BenchmarkGetPlayer(b *testing.B) {
	s := store.NewMemoryStore()
	svcs := service.NewServices(s, nil)
	h := &Handler{Services: svcs}
	server := httptest.NewServer(h.SetupHandler())
	defer server.Close()

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		resp, err := http.Get(server.URL + "/get-player?playerId=nonexistent")
		if err != nil {
			b.Fatal(err)
		}
		resp.Body.Close()
	}
}

func BenchmarkCreateRoom(b *testing.B) {
	s := store.NewMemoryStore()
	svcs := service.NewServices(s, nil)
	h := &Handler{Services: svcs}
	server := httptest.NewServer(h.SetupHandler())
	defer server.Close()

	body := []byte(`{"name":"bench-room","max-players":10}`)
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		resp, err := http.Post(server.URL+"/create-room", "application/json", bytes.NewReader(body))
		if err != nil {
			b.Fatal(err)
		}
		resp.Body.Close()
		if resp.StatusCode != http.StatusCreated {
			b.Fatalf("expected 201, got %d", resp.StatusCode)
		}
	}
}

func BenchmarkJoinRoom(b *testing.B) {
	s := store.NewMemoryStore()
	svcs := service.NewServices(s, nil)
	h := &Handler{Services: svcs}
	server := httptest.NewServer(h.SetupHandler())
	defer server.Close()

	// Pre-create room
	resp, _ := http.Post(server.URL+"/create-room", "application/json", bytes.NewReader([]byte(`{"name":"bench-room","max-players":1000}`)))
	resp.Body.Close()

	// Pre-create player
	resp, _ = http.Post(server.URL+"/create-player", "application/json", bytes.NewReader([]byte(`{"nickname":"bench"}`)))
	resp.Body.Close()

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		body := []byte(fmt.Sprintf(`{"roomId":"bench-room","playerId":"bench-p%d"}`, i))
		resp, err := http.Post(server.URL+"/join-room", "application/json", bytes.NewReader(body))
		if err != nil {
			b.Fatal(err)
		}
		resp.Body.Close()
	}
}

func BenchmarkConcurrentRequests(b *testing.B) {
	s := store.NewMemoryStore()
	svcs := service.NewServices(s, nil)
	h := &Handler{Services: svcs}
	server := httptest.NewServer(h.SetupHandler())
	defer server.Close()

	body := []byte(`{"nickname":"stress"}`)
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		var wg sync.WaitGroup
		for j := 0; j < 10; j++ {
			wg.Add(1)
			go func() {
				defer wg.Done()
				resp, err := http.Post(server.URL+"/create-player", "application/json", bytes.NewReader(body))
				if err != nil {
					b.Error(err)
					return
				}
				resp.Body.Close()
			}()
		}
		wg.Wait()
	}
}
