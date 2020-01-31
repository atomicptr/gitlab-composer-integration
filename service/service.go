package service

type Service struct {
	config Config
}

func New(config Config) *Service {
	return &Service{
		config: config,
	}
}

func (s *Service) Run() error {
	return nil
}

func (s *Service) Stop() {
}
