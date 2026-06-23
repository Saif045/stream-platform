package httpapi

import (
	"net/http"

	"stream-platform/internal/live"
)

func (s *Server) watchChannel(w http.ResponseWriter, r *http.Request) {
	slug := r.PathValue("slug")
	if slug == "" {
		writeError(w, http.StatusBadRequest, "missing channel slug")
		return
	}

	ch, err := s.channelService.GetBySlug(r.Context(), slug)
	if err != nil {
		writeError(w, http.StatusNotFound, "channel not found")
		return
	}

	stream, err := s.liveService.GetLatestStreamByChannelID(r.Context(), ch.ID)
	if err != nil {
		writeError(w, http.StatusNotFound, "no streams found for channel")
		return
	}

	mode := "vod"
	playlistURL := "/streams/" + stream.ID + "/vod/master.m3u8"

	if stream.Status == live.StreamStatusRunning {
		mode = "live"
		playlistURL = "/streams/" + stream.ID + "/live/master.m3u8"
	}

	writeWatchPage(w, stream.ID, mode, playlistURL)
}