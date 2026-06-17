package httpapi

import (
	"net/http"

	"stream-platform/internal/live"
	"stream-platform/internal/storage"
	"stream-platform/internal/vod"
)

type Server struct {
	liveService *live.Service
	vodService  *vod.Service
	store       *storage.Store
}

func NewServer(
	liveService *live.Service,
	vodService *vod.Service,
	store *storage.Store,
) *Server {
	return &Server{
		liveService: liveService,
		vodService:  vodService,
		store:       store,
	}
}

func (s *Server) Routes() http.Handler {
	mux := http.NewServeMux()

	mux.HandleFunc("POST /api/live/streams/create", s.createLiveStream)
	mux.HandleFunc("POST /api/live/streams/start", s.startLiveStream)
	mux.HandleFunc("POST /api/live/streams/stop", s.stopLiveStream)
	mux.HandleFunc("GET /api/live/streams", s.listLiveStreams)

	mux.HandleFunc("GET /api/vods", s.listVODs)
	mux.HandleFunc("GET /api/vods/{id}", s.getVOD)

	mux.HandleFunc("GET /streams/{id}/live/master.m3u8", s.getLiveMasterPlaylist)
	mux.HandleFunc("GET /streams/{id}/vod/master.m3u8", s.getVODMasterPlaylist)
	mux.HandleFunc("GET /streams/{id}/live/{quality}/playlist.m3u8", s.getLiveVariantPlaylist)
	mux.HandleFunc("GET /streams/{id}/vod/{quality}/playlist.m3u8", s.getVODVariantPlaylist)

	fileServer := http.FileServer(http.Dir(s.store.StreamsRoot()))
	mux.Handle("/streams/{id}/hls/", http.StripPrefix("/streams/", fileServer))

	return mux
}
