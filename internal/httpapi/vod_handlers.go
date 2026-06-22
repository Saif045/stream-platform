package httpapi

import "net/http"

func (s *Server) listVODs(w http.ResponseWriter, r *http.Request) {
	vods, err := s.vodService.List()
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	writeJSON(w, http.StatusOK, vods)
}

func (s *Server) getVOD(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	if id == "" {
		writeError(w, http.StatusBadRequest, "missing vod id")
		return
	}

	vod, err := s.vodService.Get(id)
	if err != nil {
		writeError(w, http.StatusNotFound, err.Error())
		return
	}

	writeJSON(w, http.StatusOK, vod)
}
