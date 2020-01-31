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

func (s *Service) createComposerRepository() (*composer.Repository, error) {
	projects, err := s.gitlabClient.FindAllComposerProjects()
	if err != nil {
		return nil, errors.Wrap(err, "could not fetch gitlab composer projects")
	}

	packages := make(map[string]composer.PackageInfo)

	for _, project := range projects {
		packages[project.Name] = s.createComposerPackageInfo(project)
	}

	composerRepository := composer.Repository{Packages: packages}
	return &composerRepository, nil
}

func (s *Service) createComposerPackageInfo(project *gitlab.ComposerProject) composer.PackageInfo {
	packageInfo := make(composer.PackageInfo)

	// add dev-master as HEAD
	packageInfo["dev-master"] = composer.VersionInfo{
		Name: project.Name,
		Source: composer.SourceInfo{
			Reference: project.Head.ID,
			Type:      "git",
			Url:       project.GitUrl(),
		},
		Type:    project.Type(),
		Version: "dev-master",
	}

	// add all project tags as well
	for _, tag := range project.Tags {
		packageInfo[tag.Name] = composer.VersionInfo{
			Name: project.Name,
			Source: composer.SourceInfo{
				Reference: tag.Commit.ID,
				Type:      "git",
				Url:       project.GitUrl(),
			},
			Type:    project.Type(),
			Version: tag.Name,
		}
	}

	return packageInfo
}

func (s *Service) Stop() error {
	err := s.httpServer.Close()
	if err != nil {
		return err
	}

	return nil
}
