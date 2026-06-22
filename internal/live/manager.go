package live

import (
	"context"
	"fmt"
	"os/exec"
	"sync"
	"time"

	"stream-platform/internal/ffmpeg"
	"stream-platform/internal/storage"
)

const rtmpBaseURL = "rtmp://localhost/live"

type Manager struct {
	runner      *ffmpeg.Runner
	store       *storage.Store
	streamStore Store

	processesMu sync.Mutex
	processes   map[string]*exec.Cmd
}

func NewManager(runner *ffmpeg.Runner, store *storage.Store, streamStore Store) *Manager {
	return &Manager{
		runner:      runner,
		store:       store,
		streamStore: streamStore,
		processes:   make(map[string]*exec.Cmd),
	}
}

func (m *Manager) StartStream(ctx context.Context, id string) error {
	stream, err := m.streamStore.GetByID(ctx, id)
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

	if err := m.streamStore.Update(ctx, stream); err != nil {
		return err
	}

	go m.waitForStream(id, cmd)

	return nil
}

func (m *Manager) StopStream(ctx context.Context, id string) error {
	stream, err := m.streamStore.GetByID(ctx, id)
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

	now := time.Now().UTC()

	stream.Status = StreamStatusStopped
	stream.Error = ""
	stream.StoppedAt = &now

	if err := m.streamStore.Update(ctx, stream); err != nil {
		return err
	}

	if err := cmd.Process.Kill(); err != nil {
		return fmt.Errorf("kill ffmpeg process: %w", err)
	}

	return nil
}

func (m *Manager) waitForStream(id string, cmd *exec.Cmd) {
	err := cmd.Wait()

	m.processesMu.Lock()
	current := m.processes[id]
	if current == cmd {
		delete(m.processes, id)
	}
	m.processesMu.Unlock()

	ctx := context.Background()

	stream, getErr := m.streamStore.GetByID(ctx, id)
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
		_ = m.streamStore.Update(ctx, stream)
		return
	}

	stream.Status = StreamStatusStopped
	stream.Error = ""
	stream.StoppedAt = &now
	_ = m.streamStore.Update(ctx, stream)
}

func (m *Manager) StartStreamByKey(ctx context.Context, streamKey string) error {
	stream, err := m.streamStore.GetByStreamKey(ctx, streamKey)
	if err != nil {
		return err
	}

	return m.StartStream(ctx, stream.ID)
}

func (m *Manager) MarkStreamDisconnectedByKey(ctx context.Context, streamKey string) error {
	stream, err := m.streamStore.GetByStreamKey(ctx, streamKey)
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

	return m.streamStore.Update(ctx, stream)
}

func (m *Manager) HydrateStream(stream *Stream) *Stream {
	return m.hydrateStream(stream)
}

func (m *Manager) hydrateStream(stream *Stream) *Stream {
	stream.RTMPURL = fmt.Sprintf("%s/%s", rtmpBaseURL, stream.StreamKey)
	stream.OutputDir = m.store.StreamDir(stream.ID)
	stream.LiveURL = "/watch/" + stream.ID + "/live"
	stream.VODURL = "/watch/" + stream.ID + "/vod"

	return stream
}
