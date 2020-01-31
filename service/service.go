package service

import (
	"fmt"
	"log"
	"net/http"
)

type Service struct {
	config      Config
	httpHandler *http.ServeMux
	httpServer  *http.Server
	logger      *log.Logger
	errorChan   chan error
}

func New(config Config, logger *log.Logger, errorChan chan error) *Service {
	handler := http.NewServeMux()
	return &Service{
		config:      config,
		httpHandler: handler,
		httpServer: &http.Server{
			Addr:              fmt.Sprintf(":%d", config.Port),
			Handler:           handler,
			ReadTimeout:       config.HttpTimeout,
			ReadHeaderTimeout: config.HttpTimeout,
		},
		logger:    logger,
		errorChan: errorChan,
	}
}

func (s *Service) Run() error {
	s.httpHandler.Handle("/", http.RedirectHandler("/packages.json", http.StatusMovedPermanently))
	s.httpHandler.HandleFunc("/packages.json", s.handlePackagesJsonEndpoint)
	return s.httpServer.ListenAndServe()
}

func (s *Service) handlePackagesJsonEndpoint(writer http.ResponseWriter, request *http.Request) {
	s.logger.Printf("Request to \"%s\" from %s (%s)", request.URL, request.UserAgent(), request.RemoteAddr)

	writer.Header().Set("Content-Type", "application/json")

	// TODO: generate packages json from gitlab
	json := `{
		"packages": {
			"test/test": {
				"dev-master": {
					"name": "test/test",
					"source": {
						"reference": "882816c7c05b5b5704e84bdb0f7ad69230df3c0c",
						"type": "git",
						"url": "git@git.domain.com:test/test.git"
					},
					"type": "project",
					"version": "dev-master"
				},
				"v1.5": {
					"name": "test/test",
					"source": {
						"reference": "882816c7c05b5b5704e84bdb0f7ad69230df3c0c",
						"type": "git",
						"url": "git@git.domain.com:test/test.git"
					},
					"type": "project",
					"version": "v1.5"
				}
			}
		}
	}`

	_, err := writer.Write([]byte(json))
	if err != nil {
		s.errorChan <- err
	}
}

func (s *Service) Stop() error {
	err := s.httpServer.Close()
	if err != nil {
		return err
	}

	return nil
}
