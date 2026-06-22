package httpapi

import (
	"encoding/json"
	"net/http"

	"stream-platform/internal/auth"
)

type createChannelRequest struct {
	Slug string `json:"slug"`
}

func (s *Server) createChannel(w http.ResponseWriter, r *http.Request) {
	userID, ok := auth.UserIDFromContext(r.Context())
	if !ok {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	var req createChannelRequest

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid json")
		return
	}

	channel, err := s.channelService.Create(r.Context(), userID, req.Slug)
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	writeJSON(w, http.StatusCreated, channel)
}

func (s *Server) listChannels(w http.ResponseWriter, r *http.Request) {
	channels, err := s.channelService.List(r.Context())
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	writeJSON(w, http.StatusOK, channels)
}

func (s *Server) listChannelStreams(w http.ResponseWriter, r *http.Request) {
	channelID := r.PathValue("id")
	if channelID == "" {
		writeError(w, http.StatusBadRequest, "missing channel id")
		return
	}

	streams, err := s.liveService.ListStreamsByChannelID(r.Context(), channelID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	writeJSON(w, http.StatusOK, streams)
}

func (s *Server) listChannelStreamsBySlug(w http.ResponseWriter, r *http.Request) {
	slug := r.PathValue("slug")
	if slug == "" {
		writeError(w, http.StatusBadRequest, "missing channel slug")
		return
	}

	channel, err := s.channelService.GetBySlug(r.Context(), slug)
	if err != nil {
		writeError(w, http.StatusNotFound, err.Error())
		return
	}

	streams, err := s.liveService.ListStreamsByChannelID(r.Context(), channel.ID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	writeJSON(w, http.StatusOK, streams)
}
