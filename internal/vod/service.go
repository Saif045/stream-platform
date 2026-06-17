package vod

import (
	"fmt"
	"os"
	"strings"

	"stream-platform/internal/storage"
)

type VODStatus string

const (
	VODStatusGrowing VODStatus = "growing"
	VODStatusReady   VODStatus = "ready"
)

type VOD struct {
	ID          string    `json:"id"`
	Status      VODStatus `json:"status"`
	PlaylistURL string    `json:"playlist_url"`
}

type Service struct {
	paths *storage.Store
}

func NewService(paths *storage.Store) *Service {
	return &Service{paths: paths}
}

func (s *Service) Get(id string) (*VOD, error) {
	hlsPlaylist := s.paths.HLSMasterPlaylist(id)

	if _, err := os.Stat(hlsPlaylist); err != nil {
		if os.IsNotExist(err) {
			return nil, fmt.Errorf("vod not found: %s", id)
		}

		return nil, err
	}

	status := VODStatusGrowing
	if playlistHasEndlist(hlsPlaylist) {
		status = VODStatusReady
	}

	return &VOD{
		ID:          id,
		Status:      status,
		PlaylistURL: "/streams/" + id + "/vod/master.m3u8",
	}, nil
}

func (s *Service) List() ([]*VOD, error) {
	streamsRoot := s.paths.StreamsRoot()

	entries, err := os.ReadDir(streamsRoot)
	if err != nil {
		if os.IsNotExist(err) {
			return []*VOD{}, nil
		}

		return nil, err
	}

	vods := make([]*VOD, 0)

	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}

		vod, err := s.Get(entry.Name())
		if err != nil {
			continue
		}

		vods = append(vods, vod)
	}

	return vods, nil
}
func playlistHasEndlist(path string) bool {
	data, err := os.ReadFile(path)
	if err != nil {
		return false
	}

	return strings.Contains(string(data), "#EXT-X-ENDLIST")
}
