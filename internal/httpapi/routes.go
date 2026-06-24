package httpapi

import (
	"net/http"
	"stream-platform/internal/auth"
)

func (s *Server) registerAuthRoutes(mux *http.ServeMux) {
	mux.HandleFunc("POST /api/auth/register", s.registerUser)
	mux.HandleFunc("POST /api/auth/login", s.loginUser)

	mux.Handle(
		"GET /api/auth/me",
		auth.Middleware(http.HandlerFunc(s.me)),
	)
}
func (s *Server) registerAPIRoutes(mux *http.ServeMux) {
	mux.Handle(
		"POST /api/channels",
		auth.Middleware(http.HandlerFunc(s.createChannel)),
	)

	mux.HandleFunc("GET /api/channels", s.listChannels)
	mux.HandleFunc("GET /api/channels/{id}/streams", s.listChannelStreams)
	mux.HandleFunc("GET /api/channels/slug/{slug}/streams", s.listChannelStreamsBySlug)

	mux.Handle(
		"POST /api/live/streams/create",
		auth.Middleware(http.HandlerFunc(s.createLiveStream)),
	)

	mux.HandleFunc("GET /api/live/streams", s.listLiveStreams)

	// Debug/manual routes. Keep for now, but these are not the normal production flow.
	mux.Handle(
		"POST /api/debug/live/streams/start",
		auth.Middleware(http.HandlerFunc(s.startLiveStream)),
	)

	mux.Handle(
		"POST /api/debug/live/streams/stop",
		auth.Middleware(http.HandlerFunc(s.stopLiveStream)),
	)

	mux.HandleFunc("GET /api/vods", s.listVODs)
	mux.HandleFunc("GET /api/vods/{id}", s.getVOD)
}

func (s *Server) registerPublicRoutes(mux *http.ServeMux) {
	mux.HandleFunc("GET /channels/{slug}/watch", s.watchChannel)
	mux.HandleFunc("GET /watch/{id}/live", s.watchLive)
	mux.HandleFunc("GET /watch/{id}/vod", s.watchVOD)

	mux.Handle(
		"GET /static/",
		http.StripPrefix("/static/", http.FileServer(http.Dir("internal/httpapi/static"))),
	)
}

func (s *Server) registerPlaybackRoutes(mux *http.ServeMux) {
	mux.HandleFunc("GET /streams/{id}/live/master.m3u8", s.getLiveMasterPlaylist)
	mux.HandleFunc("GET /streams/{id}/vod/master.m3u8", s.getVODMasterPlaylist)
	mux.HandleFunc("GET /streams/{id}/live/{quality}/playlist.m3u8", s.getLiveVariantPlaylist)
	mux.HandleFunc("GET /streams/{id}/vod/{quality}/playlist.m3u8", s.getVODVariantPlaylist)
}
func (s *Server) registerHookRoutes(mux *http.ServeMux) {
	mux.Handle(
		"POST /api/hooks/mediamtx/ready",
		requireHookSecret(s.hookSecret, http.HandlerFunc(s.mediaMTXReady)),
	)

	mux.Handle(
		"POST /api/hooks/mediamtx/not-ready",
		requireHookSecret(s.hookSecret, http.HandlerFunc(s.mediaMTXNotReady)),
	)
}
func (s *Server) registerFileRoutes(mux *http.ServeMux) {
	fileServer := http.FileServer(http.Dir(s.store.StreamsRoot()))

	mux.Handle(
		"GET /streams/{id}/hls/",
		http.StripPrefix("/streams/", fileServer),
	)
}
func (s *Server) registerHealthRoutes(mux *http.ServeMux) {
	mux.HandleFunc("GET /healthz", s.healthz)
	mux.HandleFunc("GET /readyz", s.readyz)
}