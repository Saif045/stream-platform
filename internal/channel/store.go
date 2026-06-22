package channel

import "context"

type Store interface {
	Create(ctx context.Context, channel *Channel) error
	GetByID(ctx context.Context, id string) (*Channel, error)
	List(ctx context.Context) ([]*Channel, error)
	GetBySlug(ctx context.Context, slug string) (*Channel, error)
}
