package httpapi

import (
	"encoding/json"
	"errors"
	"net/http"

	"stream-platform/internal/auth"
	"stream-platform/internal/channel"
	"stream-platform/internal/live"
)

type createLiveStreamRequest struct {
	ChannelID string `json:"channel_id"`
}

type startLiveStreamRequest struct {
	ID string `json:"id"`
}

func (s *Server) createLiveStream(w http.ResponseWriter, r *http.Request) {
	userID, ok := auth.UserIDFromContext(r.Context())
	if !ok {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	var req createLiveStreamRequest

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid json")
		return
	}

	stream, err := s.liveService.CreateStream(r.Context(), userID, req.ChannelID)
	if err != nil {
		if errors.Is(err, live.ErrForbidden) {
			writeError(w, http.StatusForbidden, "forbidden")
			return
		}

		if errors.Is(err, channel.ErrNotFound) {
			writeError(w, http.StatusNotFound, "channel not found")
			return
		}

		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	writeJSON(w, http.StatusCreated, stream)
}

func (s *Server) startLiveStream(w http.ResponseWriter, r *http.Request) {
	userID, ok := auth.UserIDFromContext(r.Context())
	if !ok {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	var req startLiveStreamRequest

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid json")
		return
	}

	if err := s.liveService.StartStream(r.Context(), userID, req.ID); err != nil {
		if errors.Is(err, live.ErrForbidden) {
			writeError(w, http.StatusForbidden, "forbidden")
			return
		}

		if errors.Is(err, channel.ErrNotFound) {
			writeError(w, http.StatusNotFound, "channel not found")
			return
		}

		writeError(w, http.StatusConflict, err.Error())
		return
	}

	writeJSON(w, http.StatusAccepted, map[string]string{
		"status": "started",
		"id":     req.ID,
	})
}

func (s *Server) stopLiveStream(w http.ResponseWriter, r *http.Request) {
	userID, ok := auth.UserIDFromContext(r.Context())
	if !ok {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	id := r.URL.Query().Get("id")
	if id == "" {
		writeError(w, http.StatusBadRequest, "missing live stream id")
		return
	}

	if err := s.liveService.StopStream(r.Context(), userID, id); err != nil {
		if errors.Is(err, live.ErrForbidden) {
			writeError(w, http.StatusForbidden, "forbidden")
			return
		}

		if errors.Is(err, channel.ErrNotFound) {
			writeError(w, http.StatusNotFound, "channel not found")
			return
		}

		writeError(w, http.StatusConflict, err.Error())
		return
	}

	writeJSON(w, http.StatusOK, map[string]string{
		"status": "stopped",
		"id":     id,
	})
}

func (s *Server) listLiveStreams(w http.ResponseWriter, r *http.Request) {
	streams, err := s.liveService.ListStreams(r.Context())
	if err != nil {
		writeInternalError(w)
		return
	}

	writeJSON(w, http.StatusOK, streams)
}
