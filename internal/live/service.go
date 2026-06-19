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
func (s *Service) StartStreamByKey(streamKey string) error {
	return s.manager.StartStreamByKey(streamKey)
}

func (s *Service) MarkStreamDisconnectedByKey(streamKey string) error {
	return s.manager.MarkStreamDisconnectedByKey(streamKey)
}
func (s *Service) GetStream(id string) (*Stream, error) {
	return s.manager.GetStream(id)
}