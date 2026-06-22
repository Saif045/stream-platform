package httpapi

import (
	"net/http"

	"stream-platform/internal/channel"
	"stream-platform/internal/live"
	"stream-platform/internal/storage"
	"stream-platform/internal/user"
	"stream-platform/internal/vod"
)

type Server struct {
	liveService    *live.Service
	vodService     *vod.Service
	channelService *channel.Service
	userService    *user.Service
	store          *storage.Store
	hookSecret     string
}

func NewServer(
	liveService *live.Service,
	vodService *vod.Service,
	channelService *channel.Service,
	userService *user.Service,
	store *storage.Store,
	hookSecret string,
) *Server {
	return &Server{
		liveService:    liveService,
		vodService:     vodService,
		channelService: channelService,
		userService:    userService,
		store:          store,
		hookSecret:     hookSecret,
	}
}
func (s *Server) Routes() http.Handler {
	mux := http.NewServeMux()

	s.registerAuthRoutes(mux)
	s.registerAPIRoutes(mux)
	s.registerPublicRoutes(mux)
	s.registerPlaybackRoutes(mux)
	s.registerHookRoutes(mux)
	s.registerFileRoutes(mux)

	return mux
}
