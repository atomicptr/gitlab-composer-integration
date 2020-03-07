package service

import (
	"fmt"
	"log"
	"net/http"

	"github.com/patrickmn/go-cache"
	"github.com/pkg/errors"

	"github.com/atomicptr/gitlab-composer-integration/gitlab"
)

type Service struct {
	config       Config
	httpHandler  *http.ServeMux
	httpServer   *http.Server
	gitlabClient *gitlab.Client
	cache        *cache.Cache
	logger       *log.Logger
	errorChan    chan error
	running      bool
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
		cache:     cache.New(config.CacheExpireDuration, cache.NoExpiration),
		logger:    logger,
		errorChan: errorChan,
	}
}

func (s *Service) Run() error {
	if err := s.gitlabClient.Validate(); err != nil {
		return errors.Wrap(err, "can't connect to gitlab")
	}

	if s.config.GitlabToken == "" {
		s.logger.Println("your Gitlab token is empty, you can only see public repositories this way")
	}

	// if no cache option is set, flush all caches...
	if s.config.NoCache {
		s.cache.Flush()
	}

	s.restoreFileCacheIfItExists()

	s.running = true
	go s.cacheUpdateHandler()

	username, password := s.config.GetHttpCredentials()

	s.httpHandler.Handle("/", http.RedirectHandler("/packages.json", http.StatusMovedPermanently))
	s.httpHandler.HandleFunc("/packages.json", basicAuth(username, password, s.handlePackagesJsonEndpoint))
	s.httpHandler.HandleFunc("/p", basicAuth(username, password, s.handleProviderEndpoint))
	s.httpHandler.HandleFunc("/notify", basicAuth(username, password, s.handleNotifyEndpoint))
	return s.httpServer.ListenAndServe()
}

func (s *Service) Stop() error {
	s.running = false
	err := s.httpServer.Close()
	if err != nil {
		return err
	}

	return nil
}
