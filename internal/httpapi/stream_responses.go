package httpapi

import (
	"time"

	"stream-platform/internal/live"
)

type publicStreamResponse struct {
	ID        string     `json:"id"`
	ChannelID string     `json:"channel_id"`
	Status    string     `json:"status"`
	CreatedAt time.Time  `json:"created_at"`
	StartedAt *time.Time `json:"started_at,omitempty"`
	StoppedAt *time.Time `json:"stopped_at,omitempty"`
	LiveURL   string     `json:"live_url,omitempty"`
	VODURL    string     `json:"vod_url,omitempty"`
}

type ownerStreamResponse struct {
	ID        string     `json:"id"`
	ChannelID string     `json:"channel_id"`
	StreamKey string     `json:"stream_key"`
	Status    string     `json:"status"`
	CreatedAt time.Time  `json:"created_at"`
	StartedAt *time.Time `json:"started_at,omitempty"`
	StoppedAt *time.Time `json:"stopped_at,omitempty"`
	RTMPURL   string     `json:"rtmp_url"`
	LiveURL   string     `json:"live_url,omitempty"`
	VODURL    string     `json:"vod_url,omitempty"`
}

func newPublicStreamResponse(stream *live.Stream) publicStreamResponse {
	return publicStreamResponse{
		ID:        stream.ID,
		ChannelID: stream.ChannelID,
		Status:    string(stream.Status),
		CreatedAt: stream.CreatedAt,
		StartedAt: stream.StartedAt,
		StoppedAt: stream.StoppedAt,
		LiveURL:   stream.LiveURL,
		VODURL:    stream.VODURL,
	}
}

func newPublicStreamResponses(streams []*live.Stream) []publicStreamResponse {
	responses := make([]publicStreamResponse, 0, len(streams))

	for _, stream := range streams {
		responses = append(responses, newPublicStreamResponse(stream))
	}

	return responses
}

func newOwnerStreamResponse(stream *live.Stream) ownerStreamResponse {
	return ownerStreamResponse{
		ID:        stream.ID,
		ChannelID: stream.ChannelID,
		StreamKey: stream.StreamKey,
		Status:    string(stream.Status),
		CreatedAt: stream.CreatedAt,
		StartedAt: stream.StartedAt,
		StoppedAt: stream.StoppedAt,
		RTMPURL:   stream.RTMPURL,
		LiveURL:   stream.LiveURL,
		VODURL:    stream.VODURL,
	}
}
