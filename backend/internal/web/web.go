package web

import (
	"context"
	"crypto/rand"
	"encoding/json"
	"fmt"
	"log"
	"log/slog"
	"net/http"
	"strconv"
	"time"

	"webgames/internal/mnkgame"

	"github.com/go-chi/chi/v5/middleware"
	"github.com/google/uuid"
	datastar "github.com/starfederation/datastar/sdk/go"
)

type contextKey string

const (
	contextKeyUserId contextKey = "userid"
)

var gamesIdCount int = 2
var games = make(map[int]*Game)

type GameStatus int

const (
	GameStatusNoOpponent GameStatus = iota
	GameStatusPlaying
	GameStatusFinished
)

type Game struct {
	ID      string
	playerX string
	playerO string
	Status  GameStatus
	g       *mnkgame.Game
}

func authMiddleware(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		cookie, err := r.Cookie("userid")
		if err != nil {
			cookie = &http.Cookie{
				Name:     "userid",
				Value:    uuid.NewString(),
				MaxAge:   86400000,
				HttpOnly: true, // Recommended: Prevents client-side JS access (security)
				Secure:   true, // Recommended: Only send over HTTPS (security)
			}

			http.SetCookie(w, cookie)
		}

		ctx := context.WithValue(r.Context(), contextKeyUserId, cookie.Value)
		r = r.WithContext(ctx)

		h.ServeHTTP(w, r)
	})
}

func ListenAndServe(addr string) error {
	mux := http.NewServeMux()
	md := func(h http.Handler) http.Handler {
		return middleware.Logger(
			middleware.Recoverer(
				authMiddleware(h),
			),
		)
	}

	assetHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Cache-Control", "no-store")
		fs := http.FileServer(http.Dir("./assets"))
		fs.ServeHTTP(w, r)
	})

	mux.Handle("GET /assets/", http.StripPrefix("/assets/", assetHandler))
	mux.Handle("GET /games/{gameId}/sse", md(sseHandler()))
	mux.Handle("GET /games/{gameId}", md(getGame()))
	mux.Handle("POST /games/{gameId}/turn", md(makeTurn()))
	mux.Handle("POST /games/{gameId}/accept", md(acceptGame()))
	mux.Handle("POST /games", md(createGame()))
	mux.Handle("GET /", md(mainHandler()))

	return http.ListenAndServe(addr, mux)
}

func mainHandler() http.Handler {
	return http.HandlerFunc(
		func(w http.ResponseWriter, r *http.Request) {
			if r.URL.Path != "/" {
				w.WriteHeader(http.StatusNotFound)
				w.Write([]byte("Not Found"))
				return
			}

			MainPage().Render(r.Context(), w)
		},
	)
}

func getGame() http.Handler {
	return http.HandlerFunc(
		func(w http.ResponseWriter, r *http.Request) {
			gameId, _ := strconv.Atoi(r.PathValue("gameId"))
			game, ok := games[gameId]
			if !ok {
				w.WriteHeader(http.StatusNotFound)
				return
			}

			GamePage(game).Render(r.Context(), w)
		},
	)
}

func createGame() http.Handler {
	return http.HandlerFunc(
		func(w http.ResponseWriter, r *http.Request) {
			game := mnkgame.NewGame(3, 3, 3)
			games[gamesIdCount] = &Game{
				playerX: "1",
				playerO: "2",
				g:       game,
			}
			gamesIdCount++

			writeJSON(w, http.StatusOK, game)
		},
	)
}

func makeTurn() http.Handler {
	return http.HandlerFunc(
		func(w http.ResponseWriter, r *http.Request) {
			gameId, _ := strconv.Atoi(r.PathValue("gameId"))
			game, ok := games[gameId]
			if !ok {
				w.WriteHeader(http.StatusNotFound)
				return
			}

			type Payload struct {
				X int
				Y int
			}

			payload, err := readJSON[Payload](r)
			if err != nil {
				w.WriteHeader(http.StatusUnprocessableEntity)
				return
			}

			log.Println(payload)

			// e := &sse.Message{}
			// e.AppendData(string(bytes))
			// sseServer.Publish(e)

			err = mnkgame.MakeTurn(game.g, mnkgame.Position{X: payload.X, Y: payload.Y})

			if err != nil {
				w.WriteHeader(http.StatusUnprocessableEntity)
				log.Println(err)
				return
			}

			sse := datastar.NewSSE(w, r)
			sse.MergeFragmentTempl(Cell(game.g.Board[payload.Y][payload.X].String(), payload.X, payload.Y))

		},
	)
}

func acceptGame() http.Handler {
	return http.HandlerFunc(
		func(w http.ResponseWriter, r *http.Request) {
			gameId, _ := strconv.Atoi(r.PathValue("gameId"))
			_, ok := games[gameId]
			if !ok {
				w.WriteHeader(http.StatusNotFound)
				return
			}

			// e := &sse.Message{}
			// e.AppendData("GAME ACCEPTED")

			// err := sseServer.Publish(e)
			// if err != nil {
			// 	fmt.Println(err)
			// }
		},
	)
}

func sseHandler() http.Handler {
	return http.HandlerFunc(
		func(w http.ResponseWriter, r *http.Request) {
			gameId, _ := strconv.Atoi(r.PathValue("gameId"))
			_, ok := games[gameId]
			if !ok {
				w.WriteHeader(http.StatusNotFound)
				return
			}

			ticker := time.NewTicker(1 * time.Second)
			defer ticker.Stop()

			sse := datastar.NewSSE(w, r)
			for {
				select {
				case <-r.Context().Done():
					slog.Debug("Client connection closed")
					return
				case <-ticker.C:
					frag := fmt.Sprintf(`<div id="random-string">%s</div>`, rand.Text())
					sse.MergeFragments(frag)
				}
			}

		},
	)
}

func writeJSON[T any](w http.ResponseWriter, status int, v T) error {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(v); err != nil {
		return fmt.Errorf("encode json: %w", err)
	}
	return nil
}

func readJSON[T any](r *http.Request) (T, error) {
	var v T
	if err := json.NewDecoder(r.Body).Decode(&v); err != nil {
		return v, fmt.Errorf("decode json: %w", err)
	}
	return v, nil
}
