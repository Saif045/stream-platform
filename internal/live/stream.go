package live

import "time"

type StreamStatus string

const (
	StreamStatusCreated StreamStatus = "created"
	StreamStatusRunning StreamStatus = "running"
	StreamStatusStopped StreamStatus = "stopped"
	StreamStatusFailed  StreamStatus = "failed"
)

type Stream struct {
	ID        string       `json:"id"`
	ChannelID string       `json:"channel_id"`
	StreamKey string       `json:"stream_key,omitempty"`
	Status    StreamStatus `json:"status"`
	Error     string       `json:"error,omitempty"`

	CreatedAt time.Time  `json:"created_at"`
	StartedAt *time.Time `json:"started_at,omitempty"`
	StoppedAt *time.Time `json:"stopped_at,omitempty"`

	RTMPURL   string `json:"rtmp_url"`
	OutputDir string `json:"output_dir"`
	LiveURL   string `json:"live_url"`
	VODURL    string `json:"vod_url"`
}
