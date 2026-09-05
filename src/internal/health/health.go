package health

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"time"
)

type Pinger interface {
	Ping(context.Context) error
}

type Server struct {
	srv *http.Server
	db  Pinger
}

func New(addr string, db Pinger) *Server {
	s := &Server{db: db}

	mux := http.NewServeMux()
	mux.HandleFunc("GET /healthz", s.live)
	mux.HandleFunc("GET /readyz", s.ready)

	s.srv = &http.Server{
		Addr:              addr,
		Handler:           mux,
		ReadHeaderTimeout: 5 * time.Second,
	}

	return s
}

func (s *Server) live(w http.ResponseWriter, _ *http.Request) {
	writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

func (s *Server) ready(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), 2*time.Second)
	defer cancel()

	if err := s.db.Ping(ctx); err != nil {
		writeJSON(w, http.StatusServiceUnavailable, map[string]string{
			"status": "unavailable",
			"db":     err.Error(),
		})

		return
	}

	writeJSON(w, http.StatusOK, map[string]string{"status": "ok", "db": "ok"})
}

func (s *Server) Run(ctx context.Context) {
	go func() {
		<-ctx.Done()

		shutdownCtx, cancel := context.WithTimeout(context.Background(), 5 * time.Second)
		defer cancel()

		if err := s.srv.Shutdown(shutdownCtx); err != nil {
			log.Printf("health: shutdown: %s", err)
		}
	}()

	if err := s.srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		log.Printf("health: serve: %s", err)
	}
}

func writeJSON(w http.ResponseWriter, code int, body any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)

	if err := json.NewEncoder(w).Encode(body); err != nil {
		log.Printf("health: encode: %s", err)
	}
}
