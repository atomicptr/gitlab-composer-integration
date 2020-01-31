package service

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"path"
	"time"

	"github.com/patrickmn/go-cache"
	"github.com/pkg/errors"

	"github.com/atomicptr/gitlab-composer-integration/composer"
	"github.com/atomicptr/gitlab-composer-integration/gitlab"
)

const cacheKey = "gitlab-composer-packages-json-cache"

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

	s.restoreFileCacheIfItExists()

	s.running = true
	go s.cacheUpdateHandler()

	s.httpHandler.Handle("/", http.RedirectHandler("/packages.json", http.StatusMovedPermanently))
	s.httpHandler.HandleFunc("/packages.json", s.handlePackagesJsonEndpoint)
	return s.httpServer.ListenAndServe()
}

func (s *Service) restoreFileCacheIfItExists() {
	cachePath := s.getCacheFilePath()

	if _, err := os.Stat(cachePath); os.IsNotExist(err) {
		s.logger.Printf("can't restore cache from file because \"%s\" does not exist.", cachePath)
		return
	}

	err := s.cache.LoadFile(cachePath)
	if err != nil {
		s.logger.Printf("could not restore cache from file because %s", err)
		return
	}

	s.logger.Printf("successfully loaded cache from file %s", cachePath)
}

func (s *Service) persistCacheInFile() {
	cachePath := s.getCacheFilePath()

	err := s.cache.SaveFile(cachePath)
	if err != nil {
		s.logger.Printf("could not persist cache in file because %s", err)
		return
	}

	s.logger.Printf("successfully persisted cache in file %s", cachePath)
}

func (s *Service) getCacheFilePath() string {
	cachePath := s.config.CacheFilePath

	if cachePath == "" {
		cachePath = path.Join(os.TempDir(), cacheKey)
	}

	return cachePath
}

func (s *Service) cacheUpdateHandler() {
	for s.running {
		_, expirationTime, found := s.cache.GetWithExpiration(cacheKey)

		// set found to false when expiration time has passed to re-cache data
		if time.Now().After(expirationTime) {
			found = false
		}

		if !found {
			s.logger.Println("no cache found (or is expired), creating new one")
			data, err := s.fetchComposerData()
			if err == nil {
				s.cache.Set(cacheKey, data, cache.DefaultExpiration)
				s.persistCacheInFile()
			} else {
				s.logger.Println(errors.Wrap(err, "could not fetch composer data"))
			}
		}

		// TODO: replace this with a ticker
		time.Sleep(30 * time.Second)
	}
}

func (s *Service) handlePackagesJsonEndpoint(writer http.ResponseWriter, request *http.Request) {
	s.logger.Printf("Request to \"%s\" from %s (%s)", request.URL, request.UserAgent(), request.RemoteAddr)

	writer.Header().Set("Content-Type", "application/json")

	// busy loop until you can get a cache...
	for {
		if content, found := s.cache.Get(cacheKey); found {
			_, err := writer.Write(content.([]byte))
			if err != nil {
				s.logger.Println(errors.Wrap(err, "could not read cache"))
			}

			return
		}

		time.Sleep(5 * time.Second)
	}
}

func (s *Service) fetchComposerData() ([]byte, error) {
	composerJson, err := s.createComposerRepository()
	if err != nil {
		return nil, errors.Wrap(err, "could not create composer repo data")
	}

	json, err := composerJson.ToJson()
	if err != nil {
		return nil, errors.Wrap(err, "could not transform data to json")
	}

	return json, nil
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
	s.running = false
	err := s.httpServer.Close()
	if err != nil {
		return err
	}

	return nil
}
