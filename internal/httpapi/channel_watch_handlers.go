package httpapi

import "net/http"

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

	if stream.Status == "running" {
		http.Redirect(w, r, stream.LiveURL, http.StatusTemporaryRedirect)
		return
	}

	http.Redirect(w, r, stream.VODURL, http.StatusTemporaryRedirect)
}
