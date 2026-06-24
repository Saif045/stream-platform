package httpapi

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"stream-platform/internal/auth"
	"stream-platform/internal/channel"
	"stream-platform/internal/live"
	"stream-platform/internal/user"
)

type fakeUserStore struct {
	createErr error
	users     map[string]*user.User
}

func (f *fakeUserStore) Create(ctx context.Context, u *user.User) error {
	if f.createErr != nil {
		return f.createErr
	}

	if f.users == nil {
		f.users = make(map[string]*user.User)
	}

	f.users[u.Username] = u
	return nil
}

func (f *fakeUserStore) GetByUsername(ctx context.Context, username string) (*user.User, error) {
	u, ok := f.users[username]
	if !ok {
		return nil, errors.New("user not found")
	}

	return u, nil
}

type fakeChannelStore struct {
	createErr error
	listErr   error
	getErr    error

	channels map[string]*channel.Channel
	bySlug   map[string]*channel.Channel
	list     []*channel.Channel
}

func (f *fakeChannelStore) Create(ctx context.Context, ch *channel.Channel) error {
	if f.createErr != nil {
		return f.createErr
	}

	if f.channels == nil {
		f.channels = make(map[string]*channel.Channel)
	}

	if f.bySlug == nil {
		f.bySlug = make(map[string]*channel.Channel)
	}

	f.channels[ch.ID] = ch
	f.bySlug[ch.Slug] = ch

	return nil
}

func (f *fakeChannelStore) GetByID(ctx context.Context, id string) (*channel.Channel, error) {
	if f.getErr != nil {
		return nil, f.getErr
	}

	ch, ok := f.channels[id]
	if !ok {
		return nil, channel.ErrNotFound
	}

	return ch, nil
}

func (f *fakeChannelStore) List(ctx context.Context) ([]*channel.Channel, error) {
	if f.listErr != nil {
		return nil, f.listErr
	}

	if f.list != nil {
		return f.list, nil
	}

	channels := make([]*channel.Channel, 0, len(f.channels))
	for _, ch := range f.channels {
		channels = append(channels, ch)
	}

	return channels, nil
}

func (f *fakeChannelStore) GetBySlug(ctx context.Context, slug string) (*channel.Channel, error) {
	ch, ok := f.bySlug[slug]
	if !ok {
		return nil, channel.ErrNotFound
	}

	return ch, nil
}

type fakeLiveStore struct {
	createErr error
	listErr   error
	getErr    error

	streams map[string]*live.Stream
	list    []*live.Stream
}

func (f *fakeLiveStore) Create(ctx context.Context, stream *live.Stream) error {
	if f.createErr != nil {
		return f.createErr
	}

	if f.streams == nil {
		f.streams = make(map[string]*live.Stream)
	}

	f.streams[stream.ID] = stream
	return nil
}

func (f *fakeLiveStore) GetByID(ctx context.Context, id string) (*live.Stream, error) {
	if f.getErr != nil {
		return nil, f.getErr
	}

	stream, ok := f.streams[id]
	if !ok {
		return nil, errors.New("stream not found")
	}

	return stream, nil
}

func (f *fakeLiveStore) GetByStreamKey(ctx context.Context, streamKey string) (*live.Stream, error) {
	for _, stream := range f.streams {
		if stream.StreamKey == streamKey {
			return stream, nil
		}
	}

	return nil, errors.New("stream not found")
}

func (f *fakeLiveStore) List(ctx context.Context) ([]*live.Stream, error) {
	if f.listErr != nil {
		return nil, f.listErr
	}

	if f.list != nil {
		return f.list, nil
	}

	streams := make([]*live.Stream, 0, len(f.streams))
	for _, stream := range f.streams {
		streams = append(streams, stream)
	}

	return streams, nil
}

func (f *fakeLiveStore) Update(ctx context.Context, stream *live.Stream) error {
	if f.streams == nil {
		f.streams = make(map[string]*live.Stream)
	}

	f.streams[stream.ID] = stream
	return nil
}

func (f *fakeLiveStore) ListByChannelID(ctx context.Context, channelID string) ([]*live.Stream, error) {
	streams := make([]*live.Stream, 0)

	for _, stream := range f.streams {
		if stream.ChannelID == channelID {
			streams = append(streams, stream)
		}
	}

	return streams, nil
}

func (f *fakeLiveStore) GetLatestByChannelID(ctx context.Context, channelID string) (*live.Stream, error) {
	for _, stream := range f.streams {
		if stream.ChannelID == channelID {
			return stream, nil
		}
	}

	return nil, errors.New("stream not found")
}

type fakeRuntime struct{}

func (f *fakeRuntime) StartStream(ctx context.Context, id string) error {
	return nil
}

func (f *fakeRuntime) StopStream(ctx context.Context, id string) error {
	return nil
}

func (f *fakeRuntime) StartStreamByKey(ctx context.Context, streamKey string) error {
	return nil
}

func (f *fakeRuntime) MarkStreamDisconnectedByKey(ctx context.Context, streamKey string) error {
	return nil
}

func (f *fakeRuntime) HydrateStream(stream *live.Stream) *live.Stream {
	stream.RTMPURL = "rtmp://localhost/live/" + stream.StreamKey
	stream.LiveURL = "/watch/" + stream.ID + "/live"
	stream.VODURL = "/watch/" + stream.ID + "/vod"
	return stream
}

func TestRegisterUserHandler(t *testing.T) {
	t.Run("invalid json returns 400", func(t *testing.T) {
		s := &Server{
			userService: user.NewService(&fakeUserStore{}),
		}

		req := httptest.NewRequest(http.MethodPost, "/api/auth/register", strings.NewReader(`{"bad"`))
		rec := httptest.NewRecorder()

		s.registerUser(rec, req)

		if rec.Code != http.StatusBadRequest {
			t.Fatalf("expected status %d, got %d", http.StatusBadRequest, rec.Code)
		}

		assertBodyContains(t, rec.Body.String(), `"error":"invalid json"`)
	})

	t.Run("duplicate username returns 409", func(t *testing.T) {
		s := &Server{
			userService: user.NewService(&fakeUserStore{
				createErr: user.ErrUsernameTaken,
			}),
		}

		req := httptest.NewRequest(
			http.MethodPost,
			"/api/auth/register",
			strings.NewReader(`{"username":"seif","password":"password123"}`),
		)
		rec := httptest.NewRecorder()

		s.registerUser(rec, req)

		if rec.Code != http.StatusConflict {
			t.Fatalf("expected status %d, got %d", http.StatusConflict, rec.Code)
		}

		assertBodyContains(t, rec.Body.String(), `"error":"username already exists"`)
	})

	t.Run("success does not expose password hash", func(t *testing.T) {
		s := &Server{
			userService: user.NewService(&fakeUserStore{}),
		}

		req := httptest.NewRequest(
			http.MethodPost,
			"/api/auth/register",
			strings.NewReader(`{"username":"seif","password":"password123"}`),
		)
		rec := httptest.NewRecorder()

		s.registerUser(rec, req)

		if rec.Code != http.StatusCreated {
			t.Fatalf("expected status %d, got %d", http.StatusCreated, rec.Code)
		}

		assertBodyNotContains(t, rec.Body.String(), "password_hash")
		assertBodyNotContains(t, rec.Body.String(), "password123")
	})
}

func TestCreateChannelHandler(t *testing.T) {
	t.Run("missing token returns 401", func(t *testing.T) {
		s := &Server{
			channelService: channel.NewService(&fakeChannelStore{}),
		}

		handler := auth.Middleware(http.HandlerFunc(s.createChannel))

		req := httptest.NewRequest(
			http.MethodPost,
			"/api/channels",
			strings.NewReader(`{"slug":"test-channel"}`),
		)
		rec := httptest.NewRecorder()

		handler.ServeHTTP(rec, req)

		if rec.Code != http.StatusUnauthorized {
			t.Fatalf("expected status %d, got %d", http.StatusUnauthorized, rec.Code)
		}

		assertBodyContains(t, rec.Body.String(), `"error":"missing token"`)
	})

	t.Run("duplicate slug returns 409", func(t *testing.T) {
		auth.SetSecret("test-secret")

		token, err := auth.GenerateToken("user-1")
		if err != nil {
			t.Fatal(err)
		}

		s := &Server{
			channelService: channel.NewService(&fakeChannelStore{
				createErr: channel.ErrSlugTaken,
			}),
		}

		handler := auth.Middleware(http.HandlerFunc(s.createChannel))

		req := httptest.NewRequest(
			http.MethodPost,
			"/api/channels",
			strings.NewReader(`{"slug":"test-channel"}`),
		)
		req.Header.Set("Authorization", "Bearer "+token)

		rec := httptest.NewRecorder()

		handler.ServeHTTP(rec, req)

		if rec.Code != http.StatusConflict {
			t.Fatalf("expected status %d, got %d", http.StatusConflict, rec.Code)
		}

		assertBodyContains(t, rec.Body.String(), `"error":"channel slug already exists"`)
	})

	t.Run("list channels does not expose user id", func(t *testing.T) {
		s := &Server{
			channelService: channel.NewService(&fakeChannelStore{
				list: []*channel.Channel{
					{
						PublicChannel: channel.PublicChannel{
							ID:        "channel-1",
							Slug:      "test-channel",
							CreatedAt: time.Now().UTC(),
						},
						UserID: "user-1",
					},
				},
			}),
		}

		req := httptest.NewRequest(http.MethodGet, "/api/channels", nil)
		rec := httptest.NewRecorder()

		s.listChannels(rec, req)

		if rec.Code != http.StatusOK {
			t.Fatalf("expected status %d, got %d", http.StatusOK, rec.Code)
		}

		assertBodyNotContains(t, rec.Body.String(), "user_id")
	})
}

func TestLiveStreamHandlers(t *testing.T) {
	t.Run("non owner create stream returns 403", func(t *testing.T) {
		auth.SetSecret("test-secret")

		token, err := auth.GenerateToken("user-2")
		if err != nil {
			t.Fatal(err)
		}

		channelStore := &fakeChannelStore{
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

		channelService := channel.NewService(channelStore)

		s := &Server{
			channelService: channelService,
			liveService: live.NewService(
				&fakeLiveStore{},
				&fakeRuntime{},
				channelService,
			),
		}

		handler := auth.Middleware(http.HandlerFunc(s.createLiveStream))

		req := httptest.NewRequest(
			http.MethodPost,
			"/api/live/streams/create",
			strings.NewReader(`{"channel_id":"channel-1"}`),
		)
		req.Header.Set("Authorization", "Bearer "+token)

		rec := httptest.NewRecorder()

		handler.ServeHTTP(rec, req)

		if rec.Code != http.StatusForbidden {
			t.Fatalf("expected status %d, got %d", http.StatusForbidden, rec.Code)
		}

		assertBodyContains(t, rec.Body.String(), `"error":"forbidden"`)
	})

	t.Run("missing channel returns 404", func(t *testing.T) {
		auth.SetSecret("test-secret")

		token, err := auth.GenerateToken("user-1")
		if err != nil {
			t.Fatal(err)
		}

		channelService := channel.NewService(&fakeChannelStore{})

		s := &Server{
			channelService: channelService,
			liveService: live.NewService(
				&fakeLiveStore{},
				&fakeRuntime{},
				channelService,
			),
		}

		handler := auth.Middleware(http.HandlerFunc(s.createLiveStream))

		req := httptest.NewRequest(
			http.MethodPost,
			"/api/live/streams/create",
			strings.NewReader(`{"channel_id":"missing-channel"}`),
		)
		req.Header.Set("Authorization", "Bearer "+token)

		rec := httptest.NewRecorder()

		handler.ServeHTTP(rec, req)

		if rec.Code != http.StatusNotFound {
			t.Fatalf("expected status %d, got %d", http.StatusNotFound, rec.Code)
		}

		assertBodyContains(t, rec.Body.String(), `"error":"channel not found"`)
	})

	t.Run("public list does not expose sensitive stream fields", func(t *testing.T) {
		s := &Server{
			liveService: live.NewService(
				&fakeLiveStore{
					list: []*live.Stream{
						{
							PublicStream: live.PublicStream{
								ID:        "stream-1",
								ChannelID: "channel-1",
								Status:    live.StreamStatusRunning,
								CreatedAt: time.Now().UTC(),
								LiveURL:   "/watch/stream-1/live",
								VODURL:    "/watch/stream-1/vod",
							},
							StreamKey: "secret-stream-key",
							RTMPURL:   "rtmp://localhost/live/secret-stream-key",
							OutputDir: "data/streams/stream-1",
							Error:     "ffmpeg failed",
						},
					},
				},
				&fakeRuntime{},
				nil,
			),
		}

		req := httptest.NewRequest(http.MethodGet, "/api/live/streams", nil)
		rec := httptest.NewRecorder()

		s.listLiveStreams(rec, req)

		if rec.Code != http.StatusOK {
			t.Fatalf("expected status %d, got %d", http.StatusOK, rec.Code)
		}

		body := rec.Body.String()

		assertBodyNotContains(t, body, "stream_key")
		assertBodyNotContains(t, body, "secret-stream-key")
		assertBodyNotContains(t, body, "rtmp_url")
		assertBodyNotContains(t, body, "output_dir")
		assertBodyNotContains(t, body, "ffmpeg failed")
	})
}

func assertBodyContains(t *testing.T, body string, want string) {
	t.Helper()

	if !strings.Contains(body, want) {
		t.Fatalf("expected body to contain %q, got %s", want, body)
	}
}

func assertBodyNotContains(t *testing.T, body string, unwanted string) {
	t.Helper()

	if strings.Contains(body, unwanted) {
		t.Fatalf("expected body not to contain %q, got %s", unwanted, body)
	}
}
