package httpapi

import (
	"encoding/json"
	"net/http"

	"stream-platform/internal/channel"
)

type CreateChannelRequest struct {
	ID        string `json:"id"`
	ChannelID string `json:"channel_id"`
	UserID    string `json:"user_id"`
	Slug      string `json:"slug"`
}

func (s *Server) createChannel(w http.ResponseWriter, r *http.Request) {
	var req CreateChannelRequest

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid json", http.StatusBadRequest)
		return
	}

	if req.ID == "" {
		http.Error(w, "missing channel id", http.StatusBadRequest)
		return
	}

	if req.UserID == "" {
		http.Error(w, "missing user id", http.StatusBadRequest)
		return
	}

	if req.Slug == "" {
		http.Error(w, "missing channel slug", http.StatusBadRequest)
		return
	}

	ch := &channel.Channel{
		ID:     req.ID,
		UserID: req.UserID,
		Slug:   req.Slug,
	}

	if err := s.channelService.Create(ch); err != nil {
		http.Error(w, err.Error(), http.StatusConflict)
		return
	}

	writeJSON(w, http.StatusCreated, ch)
}

func (s *Server) listChannels(w http.ResponseWriter, r *http.Request) {
	channels, err := s.channelService.List()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	writeJSON(w, http.StatusOK, channels)
}
func (s *Server) listChannelStreams(w http.ResponseWriter, r *http.Request) {
	channelID := r.PathValue("id")
	if channelID == "" {
		http.Error(w, "missing channel id", http.StatusBadRequest)
		return
	}

	streams, err := s.liveService.ListStreamsByChannelID(channelID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	writeJSON(w, http.StatusOK, streams)
}
func (s *Server) listChannelStreamsBySlug(w http.ResponseWriter, r *http.Request) {
	slug := r.PathValue("slug")
	if slug == "" {
		http.Error(w, "missing channel slug", http.StatusBadRequest)
		return
	}

	ch, err := s.channelService.GetBySlug(slug)
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}

	streams, err := s.liveService.ListStreamsByChannelID(ch.ID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	writeJSON(w, http.StatusOK, streams)
}
