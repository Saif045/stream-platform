package httpapi

import "net/http"

func (s *Server) watchChannel(w http.ResponseWriter, r *http.Request) {
	slug := r.PathValue("slug")
	if slug == "" {
		http.Error(w, "missing channel slug", http.StatusBadRequest)
		return
	}

	ch, err := s.channelService.GetBySlug(slug)
	if err != nil {
		http.Error(w, "channel not found", http.StatusNotFound)
		return
	}

	stream, err := s.liveService.GetLatestStreamByChannelID(ch.ID)
	if err != nil {
		http.Error(w, "no streams found for channel", http.StatusNotFound)
		return
	}

	if stream.Status == "running" {
		http.Redirect(w, r, stream.LiveURL, http.StatusTemporaryRedirect)
		return
	}

	http.Redirect(w, r, stream.VODURL, http.StatusTemporaryRedirect)
}
