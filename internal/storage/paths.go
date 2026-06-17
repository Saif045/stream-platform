package storage

import (
	"os"
	"path/filepath"
)

type Store struct {
	root string
}

func NewStore(root string) *Store {
	return &Store{root: root}
}

func (s *Store) Root() string {
	return s.root
}

func (s *Store) StreamsRoot() string {
	return filepath.Join(s.root, "streams")
}

func (s *Store) StreamDir(streamID string) string {
	return filepath.Join(s.StreamsRoot(), streamID)
}

func (s *Store) HLSDir(streamID string) string {
	return filepath.Join(s.StreamDir(streamID), "hls")
}

func (s *Store) HLSMasterPlaylist(streamID string) string {
	return filepath.Join(s.HLSDir(streamID), "master.m3u8")
}

func (s *Store) HLSVariantPlaylist(streamID string, quality string) string {
	return filepath.Join(s.HLSDir(streamID), quality, "playlist.m3u8")
}

func (s *Store) ReadHLSMasterPlaylist(streamID string) ([]byte, error) {
	return os.ReadFile(s.HLSMasterPlaylist(streamID))
}

func (s *Store) ReadHLSVariantPlaylist(streamID string, quality string) ([]byte, error) {
	return os.ReadFile(s.HLSVariantPlaylist(streamID, quality))
}
