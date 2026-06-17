package live

type Service struct {
	manager *Manager
}

func NewService(manager *Manager) *Service {
	return &Service{manager: manager}
}

func (s *Service) CreateStream(id string) (*Stream, error) {
	return s.manager.CreateStream(id)
}

func (s *Service) StartStream(id string) error {
	return s.manager.StartStream(id)
}

func (s *Service) StopStream(id string) error {
	return s.manager.StopStream(id)
}

func (s *Service) ListStreams() []*Stream {
	return s.manager.ListStreams()
}
