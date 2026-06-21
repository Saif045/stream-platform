package live

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"os/exec"
	"sync"
	"time"

	"stream-platform/internal/ffmpeg"
	"stream-platform/internal/storage"
)

const rtmpBaseURL = "rtmp://localhost/live"

type Manager struct {
	runner *ffmpeg.Runner
	store  *storage.Store
	repo   Repository

	processesMu sync.Mutex
	processes   map[string]*exec.Cmd
}

func NewManager(runner *ffmpeg.Runner, store *storage.Store, repo Repository) *Manager {
	return &Manager{
		runner:    runner,
		store:     store,
		repo:      repo,
		processes: make(map[string]*exec.Cmd),
	}
}
func (m *Manager) StartStream(id string) error {
	stream, err := m.repo.GetByID(id)
	if err != nil {
		return err
	}

	if stream.Status == StreamStatusRunning {
		return fmt.Errorf("live stream already running: %s", id)
	}

	stream = m.hydrateStream(stream)

	cmd, err := m.runner.StartLiveHLS(stream.RTMPURL, stream.OutputDir)
	if err != nil {
		return err
	}

	m.processesMu.Lock()
	m.processes[id] = cmd
	m.processesMu.Unlock()

	now := time.Now().UTC()

	stream.Status = StreamStatusRunning
	stream.Error = ""
	stream.StartedAt = &now
	stream.StoppedAt = nil

	if err := m.repo.Update(stream); err != nil {
		return err
	}

	go m.waitForStream(id, cmd)

	return nil
}

func (m *Manager) StopStream(id string) error {
	stream, err := m.repo.GetByID(id)
	if err != nil {
		return err
	}

	if stream.Status != StreamStatusRunning {
		return fmt.Errorf("live stream is not running: %s", id)
	}

	m.processesMu.Lock()
	cmd := m.processes[id]
	delete(m.processes, id)
	m.processesMu.Unlock()

	if cmd == nil || cmd.Process == nil {
		return fmt.Errorf("live stream process missing: %s", id)
	}

	if err := cmd.Process.Kill(); err != nil {
		return fmt.Errorf("kill ffmpeg process: %w", err)
	}

	now := time.Now().UTC()

	stream.Status = StreamStatusStopped
	stream.Error = ""
	stream.StoppedAt = &now

	return m.repo.Update(stream)
}

func (m *Manager) waitForStream(id string, cmd *exec.Cmd) {
	err := cmd.Wait()

	m.processesMu.Lock()
	current := m.processes[id]
	if current == cmd {
		delete(m.processes, id)
	}
	m.processesMu.Unlock()

	stream, getErr := m.repo.GetByID(id)
	if getErr != nil {
		return
	}

	if stream.Status == StreamStatusStopped {
		return
	}

	now := time.Now().UTC()

	if err != nil {
		stream.Status = StreamStatusFailed
		stream.Error = err.Error()
		stream.StoppedAt = &now
		_ = m.repo.Update(stream)
		return
	}

	stream.Status = StreamStatusStopped
	stream.Error = ""
	stream.StoppedAt = &now
	_ = m.repo.Update(stream)
}

func (m *Manager) MarkStreamDisconnectedByKey(streamKey string) error {
	stream, err := m.repo.GetByStreamKey(streamKey)
	if err != nil {
		return err
	}

	if stream.Status != StreamStatusRunning {
		return nil
	}

	now := time.Now().UTC()

	stream.Status = StreamStatusStopped
	stream.Error = ""
	stream.StoppedAt = &now

	return m.repo.Update(stream)
}

func (m *Manager) CreateStream(id string, channelID string) (*Stream, error) {
	streamKey, err := generateStreamKey()
	if err != nil {
		return nil, fmt.Errorf("generate stream key: %w", err)
	}

	stream := &Stream{
		ID:        id,
		ChannelID: channelID,
		StreamKey: streamKey,
		Status:    StreamStatusCreated,
	}

	if err := m.repo.Create(stream); err != nil {
		return nil, err
	}

	return m.hydrateStream(stream), nil
}

func (m *Manager) StartStreamByKey(streamKey string) error {
	stream, err := m.repo.GetByStreamKey(streamKey)
	if err != nil {
		return err
	}

	return m.StartStream(stream.ID)
}

func (m *Manager) GetStream(id string) (*Stream, error) {
	stream, err := m.repo.GetByID(id)
	if err != nil {
		return nil, err
	}

	return m.hydrateStream(stream), nil
}

func (m *Manager) ListStreams() []*Stream {
	streams, err := m.repo.List()
	if err != nil {
		return []*Stream{}
	}

	for _, stream := range streams {
		m.hydrateStream(stream)
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
func (m *Manager) hydrateStream(stream *Stream) *Stream {
	stream.RTMPURL = fmt.Sprintf("%s/%s", rtmpBaseURL, stream.StreamKey)
	stream.OutputDir = m.store.StreamDir(stream.ID)
	stream.LiveURL = "/watch/" + stream.ID + "/live"
	stream.VODURL = "/watch/" + stream.ID + "/vod"

	return stream
}
func (m *Manager) ListStreamsByChannelID(channelID string) ([]*Stream, error) {
	streams, err := m.repo.ListByChannelID(channelID)
	if err != nil {
		return nil, err
	}

	for _, stream := range streams {
		m.hydrateStream(stream)
	}

	return streams, nil
}
func (m *Manager) GetLatestStreamByChannelID(channelID string) (*Stream, error) {
	stream, err := m.repo.GetLatestByChannelID(channelID)
	if err != nil {
		return nil, err
	}

	return m.hydrateStream(stream), nil
}
