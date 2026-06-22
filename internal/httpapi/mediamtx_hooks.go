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
		writeError(w, http.StatusBadRequest, "invalid json")
		return
	}

	streamKey := extractStreamKey(req.Path)
	if streamKey == "" {
		writeError(w, http.StatusBadRequest, "invalid path")
		return
	}

	if err := s.liveService.StartStreamByKey(r.Context(), streamKey); err != nil {
		writeError(w, http.StatusConflict, err.Error())
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
		writeError(w, http.StatusBadRequest, "invalid json")
		return
	}

	streamKey := extractStreamKey(req.Path)
	if streamKey == "" {
		writeError(w, http.StatusBadRequest, "invalid path")
		return
	}

	if err := s.liveService.MarkStreamDisconnectedByKey(r.Context(), streamKey); err != nil {
		writeError(w, http.StatusConflict, err.Error())
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
