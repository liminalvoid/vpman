package health

import (
	"context"
	"encoding/json"
	"log/slog"
	"net/http"
	"time"
)

type Pinger interface {
	Ping(context.Context) error
}

type Server struct {
	log *slog.Logger
	srv *http.Server
	db  Pinger
}

func New(log *slog.Logger, addr string, db Pinger) *Server {
	s := &Server{log: log, db: db}

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

func (s *Server) live(w http.ResponseWriter, r *http.Request) {
	s.writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
	s.log.Debug("received health check", "from", r.RemoteAddr)
}

func (s *Server) ready(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), 2*time.Second)
	defer cancel()

	s.log.DebugContext(ctx, "received ready check", "from", r.RemoteAddr)

	if err := s.db.Ping(ctx); err != nil {
		s.writeJSON(w, http.StatusServiceUnavailable, map[string]string{
			"status": "unavailable",
			"db":     err.Error(),
		})
		s.log.ErrorContext(ctx, "DB health check failed", "err", err.Error())

		return
	}

	s.writeJSON(w, http.StatusOK, map[string]string{"status": "ok", "db": "ok"})
	s.log.DebugContext(ctx, "successful ready check", "from", r.RemoteAddr)
}

func (s *Server) Run(ctx context.Context) {
	go func() {
		<-ctx.Done()

		shutdownCtx, cancel := context.WithTimeout(context.Background(), 5 * time.Second)
		defer cancel()

		if err := s.srv.Shutdown(shutdownCtx); err != nil {
			s.log.ErrorContext(ctx, "failed shutdown", "err", err)
		}
	}()

	if err := s.srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		s.log.ErrorContext(ctx, "failed to start", "err", err)
	}
}

func (s *Server) writeJSON(w http.ResponseWriter, code int, body any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)

	if err := json.NewEncoder(w).Encode(body); err != nil {
		s.log.Error("failed to encode response", "err", err)
	}
}
