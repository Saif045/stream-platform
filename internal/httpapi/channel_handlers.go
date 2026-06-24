package httpapi

import (
	"encoding/json"
	"errors"
	"net/http"

	"stream-platform/internal/auth"
	"stream-platform/internal/channel"
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

	ch, err := s.channelService.Create(r.Context(), userID, req.Slug)
	if err != nil {
		if errors.Is(err, channel.ErrSlugTaken) {
			writeError(w, http.StatusConflict, "channel slug already exists")
			return
		}

		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	writePublic(w, http.StatusCreated, ch)
}

func (s *Server) listChannels(w http.ResponseWriter, r *http.Request) {
	channels, err := s.channelService.List(r.Context())
	if err != nil {
		writeInternalError(w)
		return
	}

	writePublicList(w, http.StatusOK, channels)
}

func (s *Server) listChannelStreams(w http.ResponseWriter, r *http.Request) {
	channelID := r.PathValue("id")
	if channelID == "" {
		writeError(w, http.StatusBadRequest, "missing channel id")
		return
	}

	streams, err := s.liveService.ListStreamsByChannelID(r.Context(), channelID)
	if err != nil {
		writeInternalError(w)
		return
	}

	writePublicList(w, http.StatusOK, streams)

}

func (s *Server) listChannelStreamsBySlug(w http.ResponseWriter, r *http.Request) {
	slug := r.PathValue("slug")
	if slug == "" {
		writeError(w, http.StatusBadRequest, "missing channel slug")
		return
	}

	ch, err := s.channelService.GetBySlug(r.Context(), slug)
	if err != nil {
		if errors.Is(err, channel.ErrNotFound) {
			writeError(w, http.StatusNotFound, "channel not found")
			return
		}

		writeInternalError(w)
		return
	}

	streams, err := s.liveService.ListStreamsByChannelID(r.Context(), ch.ID)
	if err != nil {
		writeInternalError(w)
		return
	}

	writePublicList(w, http.StatusOK, streams)

}
