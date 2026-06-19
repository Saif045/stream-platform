package httpapi

import (
	"encoding/json"
	"net/http"
	"strings"
)

type MediaMTXHookRequest struct {
	Path string `json:"path"`
}

func (s *Server) mediaMTXReady(w http.ResponseWriter, r *http.Request) {
	var req MediaMTXHookRequest

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid json", http.StatusBadRequest)
		return
	}

	streamKey := extractStreamKey(req.Path)
	if streamKey == "" {
		http.Error(w, "invalid path", http.StatusBadRequest)
		return
	}

	if err := s.liveService.StartStreamByKey(streamKey); err != nil {
		http.Error(w, err.Error(), http.StatusConflict)
		return
	}

	writeJSON(w, http.StatusOK, map[string]string{
		"status":     "started",
		"stream_key": streamKey,
	})
}

func (s *Server) mediaMTXNotReady(w http.ResponseWriter, r *http.Request) {
	var req MediaMTXHookRequest

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid json", http.StatusBadRequest)
		return
	}

	streamKey := extractStreamKey(req.Path)
	if streamKey == "" {
		http.Error(w, "invalid path", http.StatusBadRequest)
		return
	}

	if err := s.liveService.MarkStreamDisconnectedByKey(streamKey); err != nil {
		http.Error(w, err.Error(), http.StatusConflict)
		return
	}

	writeJSON(w, http.StatusOK, map[string]string{
		"status":     "stopped",
		"stream_key": streamKey,
	})
}

func extractStreamKey(path string) string {
	path = strings.TrimPrefix(path, "/")

	parts := strings.Split(path, "/")
	if len(parts) != 2 {
		return ""
	}

	if parts[0] != "live" {
		return ""
	}

	return parts[1]
}
