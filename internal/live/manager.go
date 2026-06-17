package live

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"os/exec"
	"sync"

	"stream-platform/internal/ffmpeg"
	"stream-platform/internal/storage"
)

const rtmpBaseURL = "rtmp://localhost/live"

type StreamStatus string

const (
	StreamStatusCreated StreamStatus = "created"
	StreamStatusRunning StreamStatus = "running"
	StreamStatusStopped StreamStatus = "stopped"
	StreamStatusFailed  StreamStatus = "failed"
)

type Stream struct {
	ID        string       `json:"id"`
	StreamKey string       `json:"stream_key,omitempty"`
	RTMPURL   string       `json:"rtmp_url"`
	OutputDir string       `json:"output_dir"`
	LiveURL   string       `json:"live_url"`
	VODURL    string       `json:"vod_url"`
	Status    StreamStatus `json:"status"`
	Error     string       `json:"error,omitempty"`

	cmd *exec.Cmd
}

type Manager struct {
	mu      sync.Mutex
	runner  *ffmpeg.Runner
	paths   *storage.Store
	streams map[string]*Stream
}

func NewManager(runner *ffmpeg.Runner, paths *storage.Store) *Manager {
	return &Manager{
		runner:  runner,
		paths:   paths,
		streams: make(map[string]*Stream),
	}
}

func (m *Manager) CreateStream(id string) (*Stream, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	if _, exists := m.streams[id]; exists {
		return nil, fmt.Errorf("live stream already exists: %s", id)
	}

	streamKey, err := generateStreamKey()
	if err != nil {
		return nil, fmt.Errorf("generate stream key: %w", err)
	}

	outputDir := m.paths.StreamDir(id)
	rtmpURL := fmt.Sprintf("%s/%s", rtmpBaseURL, streamKey)

	stream := &Stream{
		ID:        id,
		StreamKey: streamKey,
		RTMPURL:   rtmpURL,
		OutputDir: outputDir,
		LiveURL:   "/streams/" + id + "/live/master.m3u8",
		VODURL:    "/streams/" + id + "/vod/master.m3u8",
		Status:    StreamStatusCreated,
	}

	m.streams[id] = stream

	return stream, nil
}

func (m *Manager) StartStream(id string) error {
	m.mu.Lock()

	stream, exists := m.streams[id]
	if !exists {
		m.mu.Unlock()
		return fmt.Errorf("live stream not found: %s", id)
	}

	if stream.Status == StreamStatusRunning {
		m.mu.Unlock()
		return fmt.Errorf("live stream already running: %s", id)
	}

	cmd, err := m.runner.StartLiveHLS(stream.RTMPURL, stream.OutputDir)
	if err != nil {
		m.mu.Unlock()
		return err
	}

	stream.cmd = cmd
	stream.Status = StreamStatusRunning
	stream.Error = ""

	m.mu.Unlock()

	go m.waitForStream(stream)

	return nil
}

func (m *Manager) waitForStream(stream *Stream) {
	err := stream.cmd.Wait()

	m.mu.Lock()
	defer m.mu.Unlock()

	if stream.Status == StreamStatusStopped {
		return
	}

	if err != nil {
		stream.Status = StreamStatusFailed
		stream.Error = err.Error()
		return
	}

	stream.Status = StreamStatusStopped
}

func (m *Manager) StopStream(id string) error {
	m.mu.Lock()

	stream, exists := m.streams[id]
	if !exists {
		m.mu.Unlock()
		return fmt.Errorf("live stream not found: %s", id)
	}

	if stream.Status != StreamStatusRunning {
		m.mu.Unlock()
		return fmt.Errorf("live stream is not running: %s", id)
	}

	cmd := stream.cmd
	m.mu.Unlock()

	if cmd == nil || cmd.Process == nil {
		return fmt.Errorf("live stream process missing: %s", id)
	}

	if err := cmd.Process.Kill(); err != nil {
		return fmt.Errorf("kill ffmpeg process: %w", err)
	}

	m.mu.Lock()
	stream.Status = StreamStatusStopped
	m.mu.Unlock()

	return nil
}

func (m *Manager) ListStreams() []*Stream {
	m.mu.Lock()
	defer m.mu.Unlock()

	streams := make([]*Stream, 0, len(m.streams))
	for _, stream := range m.streams {
		streams = append(streams, stream)
	}

	return streams
}

func generateStreamKey() (string, error) {
	buf := make([]byte, 32)

	if _, err := rand.Read(buf); err != nil {
		return "", err
	}

	return hex.EncodeToString(buf), nil
}
