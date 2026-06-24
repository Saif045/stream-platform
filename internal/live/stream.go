package live

import "time"

type StreamStatus string

const (
	StreamStatusCreated StreamStatus = "created"
	StreamStatusRunning StreamStatus = "running"
	StreamStatusStopped StreamStatus = "stopped"
	StreamStatusFailed  StreamStatus = "failed"
)

type PublicStream struct {
	ID        string       `json:"id"`
	ChannelID string       `json:"channel_id"`
	Status    StreamStatus `json:"status"`

	CreatedAt time.Time  `json:"created_at"`
	StartedAt *time.Time `json:"started_at,omitempty"`
	StoppedAt *time.Time `json:"stopped_at,omitempty"`

	LiveURL string `json:"live_url,omitempty"`
	VODURL  string `json:"vod_url,omitempty"`
}

type Stream struct {
	PublicStream

	StreamKey string `json:"stream_key,omitempty"`
	Error     string `json:"error,omitempty"`
	OutputDir string `json:"-"`
	RTMPURL   string `json:"rtmp_url,omitempty"`
}

func (s *Stream) Public() any {
	if s == nil {
		return nil
	}

	return s.PublicStream
}
