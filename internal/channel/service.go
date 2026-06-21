package channel

type Service struct {
	repo Repository
}

func NewService(repo Repository) *Service {
	return &Service{repo: repo}
}

func (s *Service) Create(channel *Channel) error {
	return s.repo.Create(channel)
}

func (s *Service) GetByID(id string) (*Channel, error) {
	return s.repo.GetByID(id)
}

func (s *Service) List() ([]*Channel, error) {
	return s.repo.List()
}
func (s *Service) GetBySlug(slug string) (*Channel, error) {
	return s.repo.GetBySlug(slug)
}
