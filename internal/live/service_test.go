package live

import (
	"context"
	"errors"
	"testing"

	"stream-platform/internal/channel"
)

type fakeLiveStore struct {
	createErr error
	getErr    error
	updateErr error
	listErr   error

	hasActiveStream bool
	hasActiveErr    error

	created *Stream
	updated *Stream

	streams map[string]*Stream
	list    []*Stream
}

func (f *fakeLiveStore) Create(ctx context.Context, stream *Stream) error {
	if f.createErr != nil {
		return f.createErr
	}

	if f.streams == nil {
		f.streams = make(map[string]*Stream)
	}

	f.created = stream
	f.streams[stream.ID] = stream

	return nil
}

func (f *fakeLiveStore) GetByID(ctx context.Context, id string) (*Stream, error) {
	if f.getErr != nil {
		return nil, f.getErr
	}

	stream, ok := f.streams[id]
	if !ok {
		return nil, errors.New("stream not found")
	}

	return stream, nil
}

func (f *fakeLiveStore) GetByStreamKey(ctx context.Context, streamKey string) (*Stream, error) {
	for _, stream := range f.streams {
		if stream.StreamKey == streamKey {
			return stream, nil
		}
	}

	return nil, errors.New("stream not found")
}

func (f *fakeLiveStore) List(ctx context.Context) ([]*Stream, error) {
	var streams []*Stream

	for _, stream := range f.streams {
		streams = append(streams, stream)
	}

	return streams, nil
}

func (f *fakeLiveStore) Update(ctx context.Context, stream *Stream) error {
	if f.updateErr != nil {
		return f.updateErr
	}

	if f.streams == nil {
		f.streams = make(map[string]*Stream)
	}

	f.updated = stream
	f.streams[stream.ID] = stream

	return nil
}

func (f *fakeLiveStore) ListByChannelID(ctx context.Context, channelID string) ([]*Stream, error) {
	var streams []*Stream

	for _, stream := range f.streams {
		if stream.ChannelID == channelID {
			streams = append(streams, stream)
		}
	}

	return streams, nil
}

func (f *fakeLiveStore) GetLatestByChannelID(ctx context.Context, channelID string) (*Stream, error) {
	for _, stream := range f.streams {
		if stream.ChannelID == channelID {
			return stream, nil
		}
	}

	return nil, errors.New("stream not found")
}
func (f *fakeLiveStore) HasActiveStreamByChannelID(ctx context.Context, channelID string) (bool, error) {
	if f.hasActiveErr != nil {
		return false, f.hasActiveErr
	}

	return f.hasActiveStream, nil
}

type fakeRuntime struct {
	startedID string
	stoppedID string

	startErr error
	stopErr  error
}

func (f *fakeRuntime) StartStream(ctx context.Context, id string) error {
	if f.startErr != nil {
		return f.startErr
	}

	f.startedID = id
	return nil
}

func (f *fakeRuntime) StopStream(ctx context.Context, id string) error {
	if f.stopErr != nil {
		return f.stopErr
	}

	f.stoppedID = id
	return nil
}

func (f *fakeRuntime) StartStreamByKey(ctx context.Context, streamKey string) error {
	return nil
}

func (f *fakeRuntime) MarkStreamDisconnectedByKey(ctx context.Context, streamKey string) error {
	return nil
}

func (f *fakeRuntime) HydrateStream(stream *Stream) *Stream {
	stream.RTMPURL = "rtmp://test/live/" + stream.StreamKey
	stream.LiveURL = "/watch/" + stream.ID + "/live"
	stream.VODURL = "/watch/" + stream.ID + "/vod"

	return stream
}

type fakeChannelGetter struct {
	channels map[string]*channel.Channel
	err      error
}

func (f *fakeChannelGetter) GetByID(ctx context.Context, id string) (*channel.Channel, error) {
	if f.err != nil {
		return nil, f.err
	}

	ch, ok := f.channels[id]
	if !ok {
		return nil, errors.New("channel not found")
	}

	return ch, nil
}

func TestCreateStream(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		store := &fakeLiveStore{}
		runtime := &fakeRuntime{}
		channels := &fakeChannelGetter{
			channels: map[string]*channel.Channel{
				"channel-1": {
					PublicChannel: channel.PublicChannel{
						ID:   "channel-1",
						Slug: "test-channel",
					},
					UserID: "user-1",
				},
			},
		}

		service := NewService(store, runtime, channels)

		stream, err := service.CreateStream(context.Background(), " user-1 ", " channel-1 ")
		if err != nil {
			t.Fatal(err)
		}

		if stream.ID == "" {
			t.Fatal("expected generated stream id")
		}

		if stream.ChannelID != "channel-1" {
			t.Fatalf("expected channel id %q, got %q", "channel-1", stream.ChannelID)
		}

		if stream.StreamKey == "" {
			t.Fatal("expected generated stream key")
		}

		if stream.Status != StreamStatusCreated {
			t.Fatalf("expected status %q, got %q", StreamStatusCreated, stream.Status)
		}

		if stream.RTMPURL == "" {
			t.Fatal("expected stream to be hydrated")
		}

		if store.created != stream {
			t.Fatal("expected stream to be passed to store")
		}
	})

	t.Run("rejects missing user id", func(t *testing.T) {
		store := &fakeLiveStore{}
		runtime := &fakeRuntime{}
		channels := &fakeChannelGetter{}

		service := NewService(store, runtime, channels)

		_, err := service.CreateStream(context.Background(), "", "channel-1")
		if err == nil {
			t.Fatal("expected error")
		}

		if store.created != nil {
			t.Fatal("expected store not to be called")
		}
	})

	t.Run("rejects missing channel id", func(t *testing.T) {
		store := &fakeLiveStore{}
		runtime := &fakeRuntime{}
		channels := &fakeChannelGetter{}

		service := NewService(store, runtime, channels)

		_, err := service.CreateStream(context.Background(), "user-1", "")
		if err == nil {
			t.Fatal("expected error")
		}

		if store.created != nil {
			t.Fatal("expected store not to be called")
		}
	})

	t.Run("rejects non owner", func(t *testing.T) {
		store := &fakeLiveStore{}
		runtime := &fakeRuntime{}
		channels := &fakeChannelGetter{
			channels: map[string]*channel.Channel{
				"channel-1": {
					PublicChannel: channel.PublicChannel{
						ID:   "channel-1",
						Slug: "test-channel",
					},
					UserID: "owner-user",
				},
			},
		}

		service := NewService(store, runtime, channels)

		_, err := service.CreateStream(context.Background(), "other-user", "channel-1")
		if !errors.Is(err, ErrForbidden) {
			t.Fatalf("expected ErrForbidden, got %v", err)
		}

		if store.created != nil {
			t.Fatal("expected store not to be called")
		}
	})

	t.Run("returns channel lookup error", func(t *testing.T) {
		channelErr := errors.New("channel lookup failed")

		store := &fakeLiveStore{}
		runtime := &fakeRuntime{}
		channels := &fakeChannelGetter{err: channelErr}

		service := NewService(store, runtime, channels)

		_, err := service.CreateStream(context.Background(), "user-1", "channel-1")
		if !errors.Is(err, channelErr) {
			t.Fatalf("expected channel error, got %v", err)
		}

		if store.created != nil {
			t.Fatal("expected store not to be called")
		}
	})

	t.Run("returns store error", func(t *testing.T) {
		storeErr := errors.New("store failed")

		store := &fakeLiveStore{createErr: storeErr}
		runtime := &fakeRuntime{}
		channels := &fakeChannelGetter{
			channels: map[string]*channel.Channel{
				"channel-1": {
					PublicChannel: channel.PublicChannel{
						ID:   "channel-1",
						Slug: "test-channel",
					},
					UserID: "user-1",
				},
			},
		}

		service := NewService(store, runtime, channels)

		_, err := service.CreateStream(context.Background(), "user-1", "channel-1")
		if !errors.Is(err, storeErr) {
			t.Fatalf("expected store error, got %v", err)
		}
	})

	t.Run("rejects active stream exists", func(t *testing.T) {
		store := &fakeLiveStore{
			hasActiveStream: true,
		}

		runtime := &fakeRuntime{}

		channelGetter := &fakeChannelGetter{
			channels: map[string]*channel.Channel{
				"channel-1": {
					PublicChannel: channel.PublicChannel{
						ID: "channel-1",
					},
					UserID: "user-1",
				},
			},
		}

		service := NewService(store, runtime, channelGetter)

		stream, err := service.CreateStream(context.Background(), "user-1", "channel-1")
		if !errors.Is(err, ErrActiveStreamExists) {
			t.Fatalf("expected ErrActiveStreamExists, got %v", err)
		}

		if stream != nil {
			t.Fatalf("expected nil stream, got %+v", stream)
		}

		if store.created != nil {
			t.Fatalf("expected stream not to be created")
		}
	})

	t.Run("returns active stream check error", func(t *testing.T) {
		expectedErr := errors.New("active stream check failed")

		store := &fakeLiveStore{
			hasActiveErr: expectedErr,
		}

		runtime := &fakeRuntime{}

		channelGetter := &fakeChannelGetter{
			channels: map[string]*channel.Channel{
				"channel-1": {
					PublicChannel: channel.PublicChannel{
						ID: "channel-1",
					},
					UserID: "user-1",
				},
			},
		}

		service := NewService(store, runtime, channelGetter)

		stream, err := service.CreateStream(context.Background(), "user-1", "channel-1")
		if !errors.Is(err, expectedErr) {
			t.Fatalf("expected %v, got %v", expectedErr, err)
		}

		if stream != nil {
			t.Fatalf("expected nil stream, got %+v", stream)
		}

		if store.created != nil {
			t.Fatalf("expected stream not to be created")
		}
	})

}

func TestStartStream(t *testing.T) {
	t.Run("owner can start stream", func(t *testing.T) {
		store := &fakeLiveStore{
			streams: map[string]*Stream{
				"stream-1": {
					PublicStream: PublicStream{
						ID:        "stream-1",
						ChannelID: "channel-1",
						Status:    StreamStatusCreated,
					},
				},
			},
		}

		runtime := &fakeRuntime{}
		channels := &fakeChannelGetter{
			channels: map[string]*channel.Channel{
				"channel-1": {
					PublicChannel: channel.PublicChannel{
						ID:   "channel-1",
						Slug: "test-channel",
					},
					UserID: "user-1",
				},
			},
		}

		service := NewService(store, runtime, channels)

		err := service.StartStream(context.Background(), "user-1", "stream-1")
		if err != nil {
			t.Fatal(err)
		}

		if runtime.startedID != "stream-1" {
			t.Fatalf("expected runtime to start stream %q, got %q", "stream-1", runtime.startedID)
		}
	})

	t.Run("rejects non owner", func(t *testing.T) {
		store := &fakeLiveStore{
			streams: map[string]*Stream{
				"stream-1": {
					PublicStream: PublicStream{
						ID:        "stream-1",
						ChannelID: "channel-1",
						Status:    StreamStatusCreated,
					},
				},
			},
		}

		runtime := &fakeRuntime{}
		channels := &fakeChannelGetter{
			channels: map[string]*channel.Channel{
				"channel-1": {
					PublicChannel: channel.PublicChannel{
						ID:   "channel-1",
						Slug: "test-channel",
					},
					UserID: "owner-user",
				},
			},
		}

		service := NewService(store, runtime, channels)

		err := service.StartStream(context.Background(), "other-user", "stream-1")
		if !errors.Is(err, ErrForbidden) {
			t.Fatalf("expected ErrForbidden, got %v", err)
		}

		if runtime.startedID != "" {
			t.Fatal("expected runtime not to be called")
		}
	})

	t.Run("returns stream lookup error", func(t *testing.T) {
		storeErr := errors.New("stream lookup failed")

		store := &fakeLiveStore{getErr: storeErr}
		runtime := &fakeRuntime{}
		channels := &fakeChannelGetter{}

		service := NewService(store, runtime, channels)

		err := service.StartStream(context.Background(), "user-1", "stream-1")
		if !errors.Is(err, storeErr) {
			t.Fatalf("expected stream lookup error, got %v", err)
		}

		if runtime.startedID != "" {
			t.Fatal("expected runtime not to be called")
		}
	})

	t.Run("returns runtime error", func(t *testing.T) {
		runtimeErr := errors.New("runtime failed")

		store := &fakeLiveStore{
			streams: map[string]*Stream{
				"stream-1": {
					PublicStream: PublicStream{
						ID:        "stream-1",
						ChannelID: "channel-1",
						Status:    StreamStatusCreated,
					},
				},
			},
		}

		runtime := &fakeRuntime{startErr: runtimeErr}
		channels := &fakeChannelGetter{
			channels: map[string]*channel.Channel{
				"channel-1": {
					PublicChannel: channel.PublicChannel{
						ID:   "channel-1",
						Slug: "test-channel",
					},
					UserID: "user-1",
				},
			},
		}

		service := NewService(store, runtime, channels)

		err := service.StartStream(context.Background(), "user-1", "stream-1")
		if !errors.Is(err, runtimeErr) {
			t.Fatalf("expected runtime error, got %v", err)
		}
	})
}

func TestStopStream(t *testing.T) {
	t.Run("owner can stop stream", func(t *testing.T) {
		store := &fakeLiveStore{
			streams: map[string]*Stream{
				"stream-1": {
					PublicStream: PublicStream{
						ID:        "stream-1",
						ChannelID: "channel-1",
						Status:    StreamStatusRunning,
					},
				},
			},
		}

		runtime := &fakeRuntime{}
		channels := &fakeChannelGetter{
			channels: map[string]*channel.Channel{
				"channel-1": {
					PublicChannel: channel.PublicChannel{
						ID:   "channel-1",
						Slug: "test-channel",
					},
					UserID: "user-1",
				},
			},
		}

		service := NewService(store, runtime, channels)

		err := service.StopStream(context.Background(), "user-1", "stream-1")
		if err != nil {
			t.Fatal(err)
		}

		if runtime.stoppedID != "stream-1" {
			t.Fatalf("expected runtime to stop stream %q, got %q", "stream-1", runtime.stoppedID)
		}
	})

	t.Run("rejects non owner", func(t *testing.T) {
		store := &fakeLiveStore{
			streams: map[string]*Stream{
				"stream-1": {
					PublicStream: PublicStream{
						ID:        "stream-1",
						ChannelID: "channel-1",
						Status:    StreamStatusRunning,
					},
				},
			},
		}

		runtime := &fakeRuntime{}
		channels := &fakeChannelGetter{
			channels: map[string]*channel.Channel{
				"channel-1": {
					PublicChannel: channel.PublicChannel{
						ID:   "channel-1",
						Slug: "test-channel",
					},
					UserID: "owner-user",
				},
			},
		}

		service := NewService(store, runtime, channels)

		err := service.StopStream(context.Background(), "other-user", "stream-1")
		if !errors.Is(err, ErrForbidden) {
			t.Fatalf("expected ErrForbidden, got %v", err)
		}

		if runtime.stoppedID != "" {
			t.Fatal("expected runtime not to be called")
		}
	})

	t.Run("returns runtime error", func(t *testing.T) {
		runtimeErr := errors.New("runtime failed")

		store := &fakeLiveStore{
			streams: map[string]*Stream{
				"stream-1": {
					PublicStream: PublicStream{
						ID:        "stream-1",
						ChannelID: "channel-1",
						Status:    StreamStatusRunning,
					},
				},
			},
		}

		runtime := &fakeRuntime{stopErr: runtimeErr}
		channels := &fakeChannelGetter{
			channels: map[string]*channel.Channel{
				"channel-1": {
					PublicChannel: channel.PublicChannel{
						ID:   "channel-1",
						Slug: "test-channel",
					},
					UserID: "user-1",
				},
			},
		}

		service := NewService(store, runtime, channels)

		err := service.StopStream(context.Background(), "user-1", "stream-1")
		if !errors.Is(err, runtimeErr) {
			t.Fatalf("expected runtime error, got %v", err)
		}
	})
}
