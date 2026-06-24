package httpapi

import (
	"context"
	"net/http"
	"time"
)

func (s *Server) healthz(w http.ResponseWriter, r *http.Request) {
	writeRawJSON(w, http.StatusOK, map[string]string{
		"status": "ok",
	})
}

func (s *Server) readyz(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), 2*time.Second)
	defer cancel()

	if s.db == nil {
		writeRawJSON(w, http.StatusServiceUnavailable, map[string]any{
			"status": "not_ready",
			"checks": map[string]string{
				"database": "missing",
			},
		})
		return
	}

	if err := s.db.Ping(ctx); err != nil {
		writeRawJSON(w, http.StatusServiceUnavailable, map[string]any{
			"status": "not_ready",
			"checks": map[string]string{
				"database": "down",
			},
		})
		return
	}

	writeRawJSON(w, http.StatusOK, map[string]any{
		"status": "ready",
		"checks": map[string]string{
			"database": "ok",
		},
	})
}
