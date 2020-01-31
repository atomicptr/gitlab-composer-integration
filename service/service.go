package service

import (
	"fmt"
	"github.com/atomicptr/gitlab-composer-integration/composer"
	"github.com/atomicptr/gitlab-composer-integration/gitlab"
	"github.com/pkg/errors"
	"log"
	"net/http"
)

type Service struct {
	config       Config
	httpHandler  *http.ServeMux
	httpServer   *http.Server
	gitlabClient *gitlab.Client
	logger       *log.Logger
	errorChan    chan error
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
		gitlabClient: gitlab.New(
			config.GitlabUrl,
			config.GitlabToken,
			logger,
		),
		logger:    logger,
		errorChan: errorChan,
	}
}

func (s *Service) Run() error {
	if err := s.gitlabClient.Validate(); err != nil {
		return errors.Wrap(err, "can't connect to gitlab")
	}

	s.httpHandler.Handle("/", http.RedirectHandler("/packages.json", http.StatusMovedPermanently))
	s.httpHandler.HandleFunc("/packages.json", s.handlePackagesJsonEndpoint)
	return s.httpServer.ListenAndServe()
}

func (s *Service) handlePackagesJsonEndpoint(writer http.ResponseWriter, request *http.Request) {
	s.logger.Printf("Request to \"%s\" from %s (%s)", request.URL, request.UserAgent(), request.RemoteAddr)

	writer.Header().Set("Content-Type", "application/json")

	composerJson, err := s.createComposerRepository()
	if err != nil {
		s.errorChan <- err
		writer.WriteHeader(http.StatusInternalServerError)
		return
	}

	json, err := composerJson.ToJson()
	if err != nil {
		s.errorChan <- err
		writer.WriteHeader(http.StatusInternalServerError)
		return
	}

	_, err = writer.Write(json)
	if err != nil {
		s.errorChan <- err
		writer.WriteHeader(http.StatusInternalServerError)
		return
	}
}

func (s *Service) createComposerRepository() (composer.Repository, error) {
	return composer.Example(), nil
}

func (s *Service) Stop() error {
	err := s.httpServer.Close()
	if err != nil {
		return err
	}

	return nil
}
