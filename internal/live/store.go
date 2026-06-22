package live

import "context"

type Store interface {
	Create(ctx context.Context, stream *Stream) error
	GetByID(ctx context.Context, id string) (*Stream, error)
	GetByStreamKey(ctx context.Context, streamKey string) (*Stream, error)
	List(ctx context.Context) ([]*Stream, error)
	Update(ctx context.Context, stream *Stream) error
	ListByChannelID(ctx context.Context, channelID string) ([]*Stream, error)
	GetLatestByChannelID(ctx context.Context, channelID string) (*Stream, error)
}
