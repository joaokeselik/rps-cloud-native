package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Player struct {
	ID           int64     `json:"id"`
	Name         string    `json:"name"`
	FavoriteMove string    `json:"favorite_move"`
	Rating       int       `json:"rating"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

type playerInput struct {
	Name         string `json:"name"`
	FavoriteMove string `json:"favorite_move"`
	Rating       int    `json:"rating"`
}

type Store struct {
	pool *pgxpool.Pool
}

type Server struct {
	store *Store
}

func main() {
	ctx := context.Background()
	databaseURL := env("DATABASE_URL", "postgres://players_user:change-me@localhost:5432/players?sslmode=disable")
	port := env("PORT", "8080")

	pool, err := connectWithRetry(ctx, databaseURL, 30, 2*time.Second)
	if err != nil {
		log.Fatalf("database connection failed: %v", err)
	}
	defer pool.Close()

	store := &Store{pool: pool}
	if err := store.migrate(ctx); err != nil {
		log.Fatalf("database migration failed: %v", err)
	}

	server := &Server{store: store}
	log.Printf("players API listening on :%s", port)
	if err := http.ListenAndServe(":"+port, withCORS(server.routes())); err != nil {
		log.Fatal(err)
	}
}

func env(key, fallback string) string {
	value := strings.TrimSpace(os.Getenv(key))
	if value == "" {
		return fallback
	}
	return value
}

func connectWithRetry(ctx context.Context, databaseURL string, attempts int, delay time.Duration) (*pgxpool.Pool, error) {
	var lastErr error
	for i := 0; i < attempts; i++ {
		pool, err := pgxpool.New(ctx, databaseURL)
		if err == nil {
			pingCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
			err = pool.Ping(pingCtx)
			cancel()
			if err == nil {
				return pool, nil
			}
			pool.Close()
		}
		lastErr = err
		time.Sleep(delay)
	}
	return nil, lastErr
}

func (s *Store) migrate(ctx context.Context) error {
	_, err := s.pool.Exec(ctx, `
		CREATE TABLE IF NOT EXISTS players (
			id BIGSERIAL PRIMARY KEY,
			name TEXT NOT NULL,
			favorite_move TEXT NOT NULL,
			rating INTEGER NOT NULL DEFAULT 1000,
			created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
			updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
		)
	`)
	return err
}

func (s *Store) listPlayers(ctx context.Context) ([]Player, error) {
	rows, err := s.pool.Query(ctx, `
		SELECT id, name, favorite_move, rating, created_at, updated_at
		FROM players
		ORDER BY id DESC
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	players := make([]Player, 0)
	for rows.Next() {
		var player Player
		if err := rows.Scan(&player.ID, &player.Name, &player.FavoriteMove, &player.Rating, &player.CreatedAt, &player.UpdatedAt); err != nil {
			return nil, err
		}
		players = append(players, player)
	}
	return players, rows.Err()
}

func (s *Store) getPlayer(ctx context.Context, id int64) (Player, error) {
	var player Player
	err := s.pool.QueryRow(ctx, `
		SELECT id, name, favorite_move, rating, created_at, updated_at
		FROM players
		WHERE id = $1
	`, id).Scan(&player.ID, &player.Name, &player.FavoriteMove, &player.Rating, &player.CreatedAt, &player.UpdatedAt)
	return player, err
}

func (s *Store) createPlayer(ctx context.Context, input playerInput) (Player, error) {
	var player Player
	err := s.pool.QueryRow(ctx, `
		INSERT INTO players (name, favorite_move, rating)
		VALUES ($1, $2, $3)
		RETURNING id, name, favorite_move, rating, created_at, updated_at
	`, input.Name, input.FavoriteMove, input.Rating).Scan(&player.ID, &player.Name, &player.FavoriteMove, &player.Rating, &player.CreatedAt, &player.UpdatedAt)
	return player, err
}

func (s *Store) updatePlayer(ctx context.Context, id int64, input playerInput) (Player, error) {
	var player Player
	err := s.pool.QueryRow(ctx, `
		UPDATE players
		SET name = $2, favorite_move = $3, rating = $4, updated_at = now()
		WHERE id = $1
		RETURNING id, name, favorite_move, rating, created_at, updated_at
	`, id, input.Name, input.FavoriteMove, input.Rating).Scan(&player.ID, &player.Name, &player.FavoriteMove, &player.Rating, &player.CreatedAt, &player.UpdatedAt)
	return player, err
}

func (s *Store) deletePlayer(ctx context.Context, id int64) error {
	tag, err := s.pool.Exec(ctx, `DELETE FROM players WHERE id = $1`, id)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return pgx.ErrNoRows
	}
	return nil
}

func (s *Server) routes() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /healthz", s.health)
	mux.HandleFunc("GET /docs", docs)
	mux.HandleFunc("GET /openapi.json", openAPI)
	mux.HandleFunc("/api/players", s.players)
	mux.HandleFunc("/api/players/", s.playerByID)
	return mux
}

func (s *Server) health(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), 2*time.Second)
	defer cancel()

	if err := s.store.pool.Ping(ctx); err != nil {
		writeJSON(w, http.StatusServiceUnavailable, map[string]string{"status": "unhealthy"})
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

func (s *Server) players(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		players, err := s.store.listPlayers(r.Context())
		if err != nil {
			writeError(w, http.StatusInternalServerError, "could not list players")
			return
		}
		writeJSON(w, http.StatusOK, players)
	case http.MethodPost:
		input, ok := decodePlayerInput(w, r)
		if !ok {
			return
		}
		player, err := s.store.createPlayer(r.Context(), input)
		if err != nil {
			writeError(w, http.StatusInternalServerError, "could not create player")
			return
		}
		writeJSON(w, http.StatusCreated, player)
	default:
		w.Header().Set("Allow", "GET, POST")
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
	}
}

func (s *Server) playerByID(w http.ResponseWriter, r *http.Request) {
	id, err := parsePlayerID(r.URL.Path)
	if err != nil {
		writeError(w, http.StatusNotFound, "player not found")
		return
	}

	switch r.Method {
	case http.MethodGet:
		player, err := s.store.getPlayer(r.Context(), id)
		if errors.Is(err, pgx.ErrNoRows) {
			writeError(w, http.StatusNotFound, "player not found")
			return
		}
		if err != nil {
			writeError(w, http.StatusInternalServerError, "could not get player")
			return
		}
		writeJSON(w, http.StatusOK, player)
	case http.MethodPut:
		input, ok := decodePlayerInput(w, r)
		if !ok {
			return
		}
		player, err := s.store.updatePlayer(r.Context(), id, input)
		if errors.Is(err, pgx.ErrNoRows) {
			writeError(w, http.StatusNotFound, "player not found")
			return
		}
		if err != nil {
			writeError(w, http.StatusInternalServerError, "could not update player")
			return
		}
		writeJSON(w, http.StatusOK, player)
	case http.MethodDelete:
		if err := s.store.deletePlayer(r.Context(), id); errors.Is(err, pgx.ErrNoRows) {
			writeError(w, http.StatusNotFound, "player not found")
			return
		} else if err != nil {
			writeError(w, http.StatusInternalServerError, "could not delete player")
			return
		}
		w.WriteHeader(http.StatusNoContent)
	default:
		w.Header().Set("Allow", "GET, PUT, DELETE")
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
	}
}

func decodePlayerInput(w http.ResponseWriter, r *http.Request) (playerInput, bool) {
	var input playerInput
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		writeError(w, http.StatusBadRequest, "invalid JSON body")
		return input, false
	}
	if err := validatePlayerInput(&input); err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return input, false
	}
	return input, true
}

func validatePlayerInput(input *playerInput) error {
	input.Name = strings.TrimSpace(input.Name)
	input.FavoriteMove = strings.ToLower(strings.TrimSpace(input.FavoriteMove))

	if input.Name == "" {
		return errors.New("name is required")
	}
	switch input.FavoriteMove {
	case "rock", "paper", "scissors":
	default:
		return errors.New("favorite_move must be rock, paper, or scissors")
	}
	if input.Rating < 0 || input.Rating > 3000 {
		return errors.New("rating must be between 0 and 3000")
	}
	return nil
}

func parsePlayerID(path string) (int64, error) {
	raw := strings.TrimPrefix(path, "/api/players/")
	if raw == "" || strings.Contains(raw, "/") {
		return 0, errors.New("invalid player id")
	}
	id, err := strconv.ParseInt(raw, 10, 64)
	if err != nil || id < 1 {
		return 0, errors.New("invalid player id")
	}
	return id, nil
}

func writeJSON(w http.ResponseWriter, status int, payload any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(payload)
}

func writeError(w http.ResponseWriter, status int, message string) {
	writeJSON(w, status, map[string]string{"error": message})
}

func withCORS(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusNoContent)
			return
		}
		next.ServeHTTP(w, r)
	})
}

func docs(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	_, _ = fmt.Fprint(w, `<!doctype html>
<html lang="en">
<head>
  <meta charset="utf-8">
  <meta name="viewport" content="width=device-width, initial-scale=1">
  <title>Players API Documentation</title>
  <style>
    body { font-family: system-ui, sans-serif; margin: 2rem; color: #1f2937; line-height: 1.5; }
    code, pre { background: #f3f4f6; border-radius: 6px; padding: .2rem .35rem; }
    section { border-top: 1px solid #e5e7eb; padding-top: 1rem; margin-top: 1rem; }
  </style>
</head>
<body>
  <h1>Players CRUD API</h1>
  <p>This Go service stores player profiles in PostgreSQL.</p>
  <p>OpenAPI JSON is available at <a href="/openapi.json">/openapi.json</a>.</p>
  <section>
    <h2>Endpoints</h2>
    <ul>
      <li><code>GET /healthz</code> health check</li>
      <li><code>GET /api/players</code> list players</li>
      <li><code>POST /api/players</code> create a player</li>
      <li><code>GET /api/players/{id}</code> get one player</li>
      <li><code>PUT /api/players/{id}</code> update one player</li>
      <li><code>DELETE /api/players/{id}</code> delete one player</li>
    </ul>
  </section>
  <section>
    <h2>Example Create Request</h2>
    <pre>{
  "name": "Player One",
  "favorite_move": "rock",
  "rating": 1200
}</pre>
  </section>
</body>
</html>`)
}

func openAPI(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, map[string]any{
		"openapi": "3.0.3",
		"info": map[string]string{
			"title":   "Players CRUD API",
			"version": "1.0.0",
		},
		"paths": map[string]any{
			"/healthz": map[string]any{"get": map[string]string{"summary": "Health check"}},
			"/api/players": map[string]any{
				"get":  map[string]string{"summary": "List players"},
				"post": map[string]string{"summary": "Create player"},
			},
			"/api/players/{id}": map[string]any{
				"get":    map[string]string{"summary": "Get player"},
				"put":    map[string]string{"summary": "Update player"},
				"delete": map[string]string{"summary": "Delete player"},
			},
		},
	})
}
