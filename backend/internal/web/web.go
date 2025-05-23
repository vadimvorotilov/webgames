package web

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"

	"webgames/internal/mnkgame"
)

var gamesIdCount int = 1
var games = make(map[int]*mnkgame.Game)

func ListenAndServe(deps Deps) error {
	handler := newServer(deps)

	return http.ListenAndServe(deps.Addr, handler)
}

type Deps struct {
	Logger *log.Logger
	Addr   string
}

func newServer(deps Deps) http.Handler {
	mux := http.NewServeMux()

	mux.Handle("GET /", listGames(deps))
	mux.Handle("POST /", createGame(deps))
	mux.Handle("POST /{gameId}", makeTurn(deps))

	return mux
}

func writeJSON[T any](w http.ResponseWriter, status int, v T) error {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(v); err != nil {
		return fmt.Errorf("encode json: %w", err)
	}
	return nil
}

func listGames(deps Deps) http.Handler {
	return http.HandlerFunc(
		func(w http.ResponseWriter, r *http.Request) {
			writeJSON(w, http.StatusOK, games)
		},
	)
}

func createGame(deps Deps) http.Handler {
	return http.HandlerFunc(
		func(w http.ResponseWriter, r *http.Request) {
			game := mnkgame.NewGame(3, 3, 3)
			games[gamesIdCount] = game
			gamesIdCount++

			writeJSON(w, http.StatusOK, game)
		},
	)
}

func makeTurn(deps Deps) http.Handler {
	return http.HandlerFunc(
		func(w http.ResponseWriter, r *http.Request) {
			gameId, _ := strconv.Atoi(r.PathValue("gameId"))
			json.NewDecoder(r.Body).Decode(v any)

			if game, ok := games[gameId]; ok {
				mnkgame.MakeTurn(game, mnkgame.Position{X: 1, Y: 1})
				writeJSON(w, http.StatusOK, games)
			}

		},
	)
}
