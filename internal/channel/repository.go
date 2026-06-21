package channel

type Repository interface {
	Create(channel *Channel) error
	GetByID(id string) (*Channel, error)
	List() ([]*Channel, error)
	GetBySlug(slug string) (*Channel, error)
}
