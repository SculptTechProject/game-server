package http

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"
	"time"

	"game-server/internal/service"
	"game-server/internal/store"
)

func BenchmarkAll(b *testing.B) {
	s := store.NewMemoryStore()
	svcs := service.NewServices(s, nil)
	h := &Handler{Services: svcs}
	server := httptest.NewServer(h.SetupHandler())
	defer server.Close()

	tr := &http.Transport{
		MaxIdleConns:        1000,
		MaxIdleConnsPerHost: 1000,
		IdleConnTimeout:     60 * time.Second,
		DisableCompression:  true,
	}
	client := &http.Client{Transport: tr}
	defer tr.CloseIdleConnections()

	post := func(url string, body []byte) (*http.Response, error) {
		return client.Post(server.URL+url, "application/json", bytes.NewReader(body))
	}
	get := func(url string) (*http.Response, error) {
		return client.Get(server.URL + url)
	}
	drain := func(resp *http.Response) {
		io.Copy(io.Discard, resp.Body)
		resp.Body.Close()
	}

	body := []byte(`{"nickname":"benchmark"}`)
	roomBody := []byte(`{"name":"bench-room","max-players":1000}`)

	b.Run("CreatePlayer", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			resp, err := post("/create-player", body)
			if err != nil {
				b.Fatal(err)
			}
			drain(resp)
			if resp.StatusCode != http.StatusCreated {
				b.Fatalf("expected 201, got %d", resp.StatusCode)
			}
		}
	})

	b.Run("GetPlayer", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			resp, err := get("/get-player?playerId=nonexistent")
			if err != nil {
				b.Fatal(err)
			}
			drain(resp)
		}
	})

	b.Run("CreateRoom", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			resp, err := post("/create-room", roomBody)
			if err != nil {
				b.Fatal(err)
			}
			drain(resp)
			if resp.StatusCode != http.StatusCreated {
				b.Fatalf("expected 201, got %d", resp.StatusCode)
			}
		}
	})

	b.Run("JoinRoom", func(b *testing.B) {
		resp, _ := post("/create-room", []byte(`{"name":"bench-room2","max-players":1000}`))
		drain(resp)
		resp, _ = post("/create-player", []byte(`{"nickname":"bench"}`))
		drain(resp)

		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			jbody := []byte(fmt.Sprintf(`{"roomId":"bench-room2","playerId":"bench-p%d"}`, i))
			resp, err := post("/join-room", jbody)
			if err != nil {
				b.Fatal(err)
			}
			drain(resp)
		}
	})

	b.Run("Concurrent", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			var wg sync.WaitGroup
			for j := 0; j < 10; j++ {
				wg.Add(1)
				go func() {
					defer wg.Done()
					resp, err := post("/create-player", body)
					if err != nil {
						b.Error(err)
						return
					}
					drain(resp)
				}()
			}
			wg.Wait()
		}
	})
}
