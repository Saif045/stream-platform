package httpapi

import (
	"encoding/json"
	"net/http"
)

type CreateLiveStreamRequest struct {
	ID string `json:"id"`
}

type StartLiveStreamRequest struct {
	ID string `json:"id"`
}

func (s *Server) createLiveStream(w http.ResponseWriter, r *http.Request) {
	var req CreateLiveStreamRequest

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid json", http.StatusBadRequest)
		return
	}

	if req.ID == "" {
		http.Error(w, "missing live stream id", http.StatusBadRequest)
		return
	}

	stream, err := s.liveService.CreateStream(req.ID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusConflict)
		return
	}

	writeJSON(w, http.StatusCreated, stream)
}

func (s *Server) startLiveStream(w http.ResponseWriter, r *http.Request) {
	var req StartLiveStreamRequest

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid json", http.StatusBadRequest)
		return
	}

	if req.ID == "" {
		http.Error(w, "missing live stream id", http.StatusBadRequest)
		return
	}

	if err := s.liveService.StartStream(req.ID); err != nil {
		http.Error(w, err.Error(), http.StatusConflict)
		return
	}

	writeJSON(w, http.StatusAccepted, map[string]string{
		"status": "started",
		"id":     req.ID,
	})
}

func (s *Server) stopLiveStream(w http.ResponseWriter, r *http.Request) {
	id := r.URL.Query().Get("id")
	if id == "" {
		http.Error(w, "missing live stream id", http.StatusBadRequest)
		return
	}

	if err := s.liveService.StopStream(id); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	writeJSON(w, http.StatusOK, map[string]string{
		"status": "stopped",
		"id":     id,
	})
}

func (s *Server) listLiveStreams(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, s.liveService.ListStreams())
}
