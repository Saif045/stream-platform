package channel

import (
	"context"
	"errors"
	"regexp"
	"strings"

	"github.com/google/uuid"
)

var slugPattern = regexp.MustCompile(`^[a-z0-9-]{3,32}$`)

type Service struct {
	store Store
}

func NewService(store Store) *Service {
	return &Service{store: store}
}

func (s *Service) Create(ctx context.Context, userID string, slug string) (*Channel, error) {
	userID = strings.TrimSpace(userID)
	slug = strings.TrimSpace(slug)

	if userID == "" {
		return nil, errors.New("user id is required")
	}

	if slug == "" {
		return nil, errors.New("channel slug is required")
	}

	if !slugPattern.MatchString(slug) {
		return nil, errors.New("invalid channel slug")
	}

	channel := &Channel{
		ID:     uuid.NewString(),
		UserID: userID,
		Slug:   slug,
	}

	if err := s.store.Create(ctx, channel); err != nil {
		return nil, err
	}

	return channel, nil
}

func (s *Service) GetByID(ctx context.Context, id string) (*Channel, error) {
	return s.store.GetByID(ctx, id)
}

func (s *Service) List(ctx context.Context) ([]*Channel, error) {
	return s.store.List(ctx)
}

func (s *Service) GetBySlug(ctx context.Context, slug string) (*Channel, error) {
	return s.store.GetBySlug(ctx, slug)
}
