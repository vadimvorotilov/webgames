package web

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"log/slog"
	"net/http"
	"time"
	"webgames/internal/mnkgame"
	"webgames/internal/pubsub"

	"github.com/go-chi/chi/v5/middleware"
	"github.com/google/uuid"
	datastar "github.com/starfederation/datastar/sdk/go"
)

type contextKey string

const (
	contextKeyUserID contextKey = "userid"
)

func ListenAndServe(addr string) error {

	mux := http.NewServeMux()
	md := func(h http.Handler) http.Handler {
		return middleware.Logger(
			middleware.Recoverer(
				authMiddleware(h),
			),
		)
	}

	ps := pubsub.NewPubSub[struct{}]()

	assetHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Cache-Control", "no-store")
		fs := http.FileServer(http.Dir("./assets"))
		fs.ServeHTTP(w, r)
	})

	mux.Handle("GET /assets/", http.StripPrefix("/assets/", assetHandler))
	mux.Handle("GET /games/{gameID}/sse", md(sseHandler(ps)))
	mux.Handle("GET /games/{gameID}", md(getGame()))
	mux.Handle("POST /games/{gameID}/turn", md(makeTurn(ps)))
	mux.Handle("POST /games/{gameID}/opponent", md(becomeOpponent(ps)))
	mux.Handle("POST /games", md(createGame()))
	mux.Handle("GET /", md(mainHandler()))

	server := &http.Server{
		Addr:              addr,
		Handler:           mux,
		ReadTimeout:       10 * time.Second,
		ReadHeaderTimeout: 10 * time.Second,
		WriteTimeout:      0,
	}

	return server.ListenAndServe()
}

func mainHandler() http.Handler {
	return http.HandlerFunc(
		func(w http.ResponseWriter, r *http.Request) {
			if r.URL.Path != "/" {
				w.WriteHeader(http.StatusNotFound)
				return
			}

			MainPage().Render(r.Context(), w)
		},
	)
}

func getGame() http.Handler {
	return http.HandlerFunc(
		func(w http.ResponseWriter, r *http.Request) {
			gameID := mnkgame.GameID(r.PathValue("gameID"))
			game, ok := mnkgame.FindGame(gameID)
			if !ok {
				w.WriteHeader(http.StatusNotFound)
				return
			}
			playerID := r.Context().Value(contextKeyUserID).(string)

			GamePage(game, mnkgame.PlayerID(playerID)).Render(r.Context(), w)
		},
	)
}

func getUserID(ctx context.Context) string {
	return ctx.Value(contextKeyUserID).(string)
}

func createGame() http.Handler {
	return http.HandlerFunc(
		func(w http.ResponseWriter, r *http.Request) {
			params, err := readJSON[mnkgame.CreateGameParams](r)
			if err != nil {
				w.WriteHeader(http.StatusUnprocessableEntity)
				log.Println(err)
				return
			}

			params.PlayerXID = getUserID(r.Context())
			game := mnkgame.CreateGame(params)

			log.Println(*game)

			sse := datastar.NewSSE(w, r)
			url := fmt.Sprintf("/games/%s", game.ID)
			sse.Redirect(url)
		},
	)
}

func makeTurn(ps *pubsub.PubSub[struct{}]) http.Handler {
	return http.HandlerFunc(
		func(w http.ResponseWriter, r *http.Request) {
			gameID := mnkgame.GameID(r.PathValue("gameID"))
			game, ok := mnkgame.FindGame(gameID)
			if !ok {
				w.WriteHeader(http.StatusNotFound)
				return
			}

			position, err := readJSON[mnkgame.Position](r)
			if err != nil {
				w.WriteHeader(http.StatusUnprocessableEntity)
				return
			}

			err = mnkgame.MakeTurn(game, position)

			if err != nil {
				w.WriteHeader(http.StatusUnprocessableEntity)
				log.Println(err)
				return
			}

			// sse := datastar.NewSSE(w, r)
			// sse.MergeFragmentTempl(Cell(game.Board[position.Y][position.X].String(), position.X, position.Y))

			ps.Publish("game", struct{}{})
		},
	)
}

func becomeOpponent(ps *pubsub.PubSub[struct{}]) http.Handler {
	return http.HandlerFunc(
		func(w http.ResponseWriter, r *http.Request) {
			gameID := mnkgame.GameID(r.PathValue("gameID"))
			game, ok := mnkgame.FindGame(gameID)
			if !ok {
				w.WriteHeader(http.StatusNotFound)
				return
			}

			playerID := r.Context().Value(contextKeyUserID).(string)
			err := mnkgame.BecomeOpponent(game, mnkgame.PlayerID(playerID))
			if err != nil {
				w.WriteHeader(http.StatusBadRequest)
				return
			}

			ps.Publish("game", struct{}{})

			// sse := datastar.NewSSE(w, r)
			// sse.MergeFragmentTempl(GameBoard(game, mnkgame.PlayerID(playerID)))
		},
	)
}

func sseHandler(ps *pubsub.PubSub[struct{}]) http.Handler {
	return http.HandlerFunc(
		func(w http.ResponseWriter, r *http.Request) {
			gameID := mnkgame.GameID(r.PathValue("gameID"))
			_, ok := mnkgame.FindGame(gameID)
			if !ok {
				w.WriteHeader(http.StatusNotFound)
				return
			}

			playerID := r.Context().Value(contextKeyUserID).(string)
			ch := ps.Subscribe("game")
			defer ps.Unsubscribe("game", ch)

			sse := datastar.NewSSE(w, r)
			for {
				select {
				case <-r.Context().Done():
					slog.Debug("Client connection closed")
					return
				case <-ch:
					game, _ := mnkgame.FindGame(gameID)
					sse.MergeFragmentTempl(GameBoard(game, mnkgame.PlayerID(playerID)))
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

// userid := r.Context().Value(contextKeyUserID).(string)
func authMiddleware(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		cookie, err := r.Cookie(string(contextKeyUserID))
		if err != nil {
			cookie = &http.Cookie{
				Name:     string(contextKeyUserID),
				Value:    uuid.NewString(),
				Path:     "/",
				MaxAge:   86400000,
				HttpOnly: true, // Recommended: Prevents client-side JS access (security)
				Secure:   true, // Recommended: Only send over HTTPS (security)
			}

			http.SetCookie(w, cookie)
		}

		ctx := context.WithValue(r.Context(), contextKeyUserID, cookie.Value)
		r = r.WithContext(ctx)

		h.ServeHTTP(w, r)
	})
}
