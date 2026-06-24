package live

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"strings"

	"github.com/google/uuid"

	"stream-platform/internal/channel"
)

var ErrForbidden = errors.New("forbidden")

type ChannelGetter interface {
	GetByID(ctx context.Context, id string) (*channel.Channel, error)
}

type Runtime interface {
	StartStream(ctx context.Context, id string) error
	StopStream(ctx context.Context, id string) error
	StartStreamByKey(ctx context.Context, streamKey string) error
	MarkStreamDisconnectedByKey(ctx context.Context, streamKey string) error
	HydrateStream(stream *Stream) *Stream
}
type Service struct {
	store          Store
	runtime        Runtime
	channelService ChannelGetter
}

func NewService(store Store, runtime Runtime, channelService ChannelGetter) *Service {
	return &Service{
		store:          store,
		runtime:        runtime,
		channelService: channelService,
	}
}

func (s *Service) CreateStream(ctx context.Context, userID string, channelID string) (*Stream, error) {
	userID = strings.TrimSpace(userID)
	channelID = strings.TrimSpace(channelID)

	if userID == "" {
		return nil, errors.New("user id is required")
	}

	if channelID == "" {
		return nil, errors.New("channel id is required")
	}

	ch, err := s.channelService.GetByID(ctx, channelID)
	if err != nil {
		return nil, err
	}

	if ch.UserID != userID {
		return nil, ErrForbidden
	}

	streamKey, err := generateStreamKey()
	if err != nil {
		return nil, err
	}

	stream := &Stream{
		PublicStream: PublicStream{
			ID:        uuid.NewString(),
			ChannelID: channelID,
			Status:    StreamStatusCreated,
		},
		StreamKey: streamKey,
	}
	if err := s.store.Create(ctx, stream); err != nil {
		return nil, err
	}

	return s.runtime.HydrateStream(stream), nil
}

func (s *Service) StartStream(ctx context.Context, userID string, id string) error {
	if err := s.checkStreamOwnership(ctx, userID, id); err != nil {
		return err
	}

	return s.runtime.StartStream(ctx, id)
}

func (s *Service) StopStream(ctx context.Context, userID string, id string) error {
	if err := s.checkStreamOwnership(ctx, userID, id); err != nil {
		return err
	}

	return s.runtime.StopStream(ctx, id)
}

func (s *Service) ListStreams(ctx context.Context) ([]*Stream, error) {
	streams, err := s.store.List(ctx)
	if err != nil {
		return nil, err
	}

	for _, stream := range streams {
		s.runtime.HydrateStream(stream)
	}

	return streams, nil
}

func (s *Service) StartStreamByKey(ctx context.Context, streamKey string) error {
	return s.runtime.StartStreamByKey(ctx, streamKey)
}

func (s *Service) MarkStreamDisconnectedByKey(ctx context.Context, streamKey string) error {
	return s.runtime.MarkStreamDisconnectedByKey(ctx, streamKey)
}

func (s *Service) GetStream(ctx context.Context, id string) (*Stream, error) {
	stream, err := s.store.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	return s.runtime.HydrateStream(stream), nil
}

func (s *Service) ListStreamsByChannelID(ctx context.Context, channelID string) ([]*Stream, error) {
	streams, err := s.store.ListByChannelID(ctx, channelID)
	if err != nil {
		return nil, err
	}

	for _, stream := range streams {
		s.runtime.HydrateStream(stream)
	}

	return streams, nil
}

func (s *Service) GetLatestStreamByChannelID(ctx context.Context, channelID string) (*Stream, error) {
	stream, err := s.store.GetLatestByChannelID(ctx, channelID)
	if err != nil {
		return nil, err
	}

	return s.runtime.HydrateStream(stream), nil
}

func (s *Service) checkStreamOwnership(ctx context.Context, userID string, streamID string) error {
	stream, err := s.store.GetByID(ctx, streamID)
	if err != nil {
		return err
	}

	ch, err := s.channelService.GetByID(ctx, stream.ChannelID)
	if err != nil {
		return err
	}

	if ch.UserID != userID {
		return ErrForbidden
	}

	return nil
}

func generateStreamKey() (string, error) {
	buf := make([]byte, 32)

	if _, err := rand.Read(buf); err != nil {
		return "", err
	}

	return hex.EncodeToString(buf), nil
}
