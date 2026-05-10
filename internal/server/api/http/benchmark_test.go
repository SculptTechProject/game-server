package http

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"
	"time"

	"game-server/internal/server/domain"
	"game-server/internal/service"
	"game-server/internal/store"
)

type benchEnv struct {
	server *httptest.Server
	client *http.Client
	body   []byte
	room   []byte
}

func newBenchEnv() *benchEnv {
	s := store.NewMemoryStore()
	svcs := service.NewServices(s, nil)
	h := &Handler{Services: svcs}
	server := httptest.NewServer(h.SetupHandler())

	tr := &http.Transport{
		MaxIdleConns:        1000,
		MaxIdleConnsPerHost: 1000,
		IdleConnTimeout:     60 * time.Second,
		DisableCompression:  true,
	}
	client := &http.Client{Transport: tr}

	return &benchEnv{
		server: server,
		client: client,
		body:   []byte(`{"nickname":"benchmark"}`),
		room:   []byte(`{"name":"bench-room","max-players":1000}`),
	}
}

func (e *benchEnv) close() {
	e.client.Transport.(*http.Transport).CloseIdleConnections()
	e.server.Close()
}

func (e *benchEnv) post(url string, body []byte) (*http.Response, error) {
	return e.client.Post(e.server.URL+url, "application/json", bytes.NewReader(body))
}

func (e *benchEnv) get(url string) (*http.Response, error) {
	return e.client.Get(e.server.URL + url)
}

func drain(resp *http.Response) {
	io.Copy(io.Discard, resp.Body)
	resp.Body.Close()
}

func BenchmarkAll(b *testing.B) {
	b.Run("CreatePlayer", func(b *testing.B) {
		e := newBenchEnv()
		defer e.close()
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			resp, err := e.post("/create-player", e.body)
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
		e := newBenchEnv()
		defer e.close()
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			resp, err := e.get("/get-player?playerId=nonexistent")
			if err != nil {
				b.Fatal(err)
			}
			drain(resp)
		}
	})

	b.Run("CreateRoom", func(b *testing.B) {
		e := newBenchEnv()
		defer e.close()
		// Limit iterations to stay within the ~9000 room code space
		maxOps := 4000
		if b.N < maxOps {
			maxOps = b.N
		}
		b.ResetTimer()
		for i := 0; i < maxOps; i++ {
			resp, err := e.post("/create-room", e.room)
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
		e := newBenchEnv()
		defer e.close()

		resp, _ := e.post("/create-room", []byte(`{"name":"bench-room2","max-players":1000}`))
		var roomResp domain.Room
		json.NewDecoder(resp.Body).Decode(&roomResp)
		resp.Body.Close()
		roomCode := roomResp.ID

		resp, _ = e.post("/create-player", []byte(`{"nickname":"bench"}`))
		drain(resp)

		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			jbody := []byte(fmt.Sprintf(`{"roomId":"%s","playerId":"bench-p%d"}`, roomCode, i))
			resp, err := e.post("/join-room", jbody)
			if err != nil {
				b.Fatal(err)
			}
			drain(resp)
		}
	})

	b.Run("Concurrent", func(b *testing.B) {
		e := newBenchEnv()
		defer e.close()
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			var wg sync.WaitGroup
			for j := 0; j < 10; j++ {
				wg.Add(1)
				go func() {
					defer wg.Done()
					resp, err := e.post("/create-player", e.body)
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
