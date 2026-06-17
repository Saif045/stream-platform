package httpapi

import "net/http"

func (s *Server) listVODs(w http.ResponseWriter, r *http.Request) {
	vods, err := s.vodService.List()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	writeJSON(w, http.StatusOK, vods)
}

func (s *Server) getVOD(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	if id == "" {
		http.Error(w, "missing vod id", http.StatusBadRequest)
		return
	}

	vod, err := s.vodService.Get(id)
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}

	writeJSON(w, http.StatusOK, vod)
}
