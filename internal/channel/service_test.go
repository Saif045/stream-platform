package channel

import (
	"context"
	"errors"
	"testing"
)

type fakeStore struct {
	created *Channel
	err     error
}

func (f *fakeStore) Create(ctx context.Context, ch *Channel) error {
	if f.err != nil {
		return f.err
	}

	f.created = ch
	return nil
}

func (f *fakeStore) GetByID(ctx context.Context, id string) (*Channel, error) {
	return nil, nil
}

func (f *fakeStore) List(ctx context.Context) ([]*Channel, error) {
	return nil, nil
}

func (f *fakeStore) GetBySlug(ctx context.Context, slug string) (*Channel, error) {
	return nil, nil
}

func TestCreate(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		store := &fakeStore{}
		service := NewService(store)

		ch, err := service.Create(context.Background(), " user-1 ", " test-channel ")
		if err != nil {
			t.Fatal(err)
		}

		if ch.ID == "" {
			t.Fatal("expected generated channel id")
		}

		if ch.UserID != "user-1" {
			t.Fatalf("expected user id %q, got %q", "user-1", ch.UserID)
		}

		if ch.Slug != "test-channel" {
			t.Fatalf("expected slug %q, got %q", "test-channel", ch.Slug)
		}

		if store.created != ch {
			t.Fatal("expected created channel to be passed to store")
		}
	})

	t.Run("rejects missing user id", func(t *testing.T) {
		store := &fakeStore{}
		service := NewService(store)

		_, err := service.Create(context.Background(), "", "test-channel")
		if err == nil {
			t.Fatal("expected error")
		}

		if store.created != nil {
			t.Fatal("expected store not to be called")
		}
	})

	t.Run("rejects missing slug", func(t *testing.T) {
		store := &fakeStore{}
		service := NewService(store)

		_, err := service.Create(context.Background(), "user-1", "")
		if err == nil {
			t.Fatal("expected error")
		}

		if store.created != nil {
			t.Fatal("expected store not to be called")
		}
	})

	t.Run("rejects invalid slug", func(t *testing.T) {
		store := &fakeStore{}
		service := NewService(store)

		_, err := service.Create(context.Background(), "user-1", "https://evil.com")
		if err == nil {
			t.Fatal("expected error")
		}

		if store.created != nil {
			t.Fatal("expected store not to be called")
		}
	})

	t.Run("returns store error", func(t *testing.T) {
		storeErr := errors.New("store failed")
		store := &fakeStore{err: storeErr}
		service := NewService(store)

		_, err := service.Create(context.Background(), "user-1", "test-channel")
		if !errors.Is(err, storeErr) {
			t.Fatalf("expected store error, got %v", err)
		}
	})
}
