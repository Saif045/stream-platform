package live

import (
	"fmt"
	"sync"
)

type MemoryRepository struct {
	mu      sync.RWMutex
	streams map[string]*Stream
}

func NewMemoryRepository() *MemoryRepository {
	return &MemoryRepository{
		streams: make(map[string]*Stream),
	}
}

func (r *MemoryRepository) Create(stream *Stream) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if _, exists := r.streams[stream.ID]; exists {
		return fmt.Errorf("live stream already exists: %s", stream.ID)
	}

	copied := *stream
	r.streams[stream.ID] = &copied

	return nil
}

func (r *MemoryRepository) GetByID(id string) (*Stream, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	stream, exists := r.streams[id]
	if !exists {
		return nil, fmt.Errorf("live stream not found: %s", id)
	}

	copied := *stream
	return &copied, nil
}

func (r *MemoryRepository) GetByStreamKey(streamKey string) (*Stream, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	for _, stream := range r.streams {
		if stream.StreamKey == streamKey {
			copied := *stream
			return &copied, nil
		}
	}

	return nil, fmt.Errorf("stream key not found")
}

func (r *MemoryRepository) List() ([]*Stream, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	streams := make([]*Stream, 0, len(r.streams))

	for _, stream := range r.streams {
		copied := *stream
		streams = append(streams, &copied)
	}

	return streams, nil
}

func (r *MemoryRepository) Update(stream *Stream) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if _, exists := r.streams[stream.ID]; !exists {
		return fmt.Errorf("live stream not found: %s", stream.ID)
	}

	copied := *stream
	r.streams[stream.ID] = &copied

	return nil
}
