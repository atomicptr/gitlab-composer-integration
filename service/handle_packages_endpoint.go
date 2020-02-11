package service

import (
	"net/http"
	"time"

	"github.com/pkg/errors"

	"github.com/atomicptr/gitlab-composer-integration/composer"
	"github.com/atomicptr/gitlab-composer-integration/gitlab"
)

func (s *Service) handlePackagesJsonEndpoint(writer http.ResponseWriter, request *http.Request) {
	s.logger.Printf("Request to \"%s\" from %s (%s)\n", request.URL, request.UserAgent(), request.RemoteAddr)

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

	jsonData, err := composerJson.ToJson()
	if err != nil {
		return nil, errors.Wrap(err, "could not transform data to json")
	}

	return jsonData, nil
}

func (s *Service) createComposerRepository() (*composer.Repository, error) {
	projects, err := s.gitlabClient.FindAllComposerProjects()
	if err != nil {
		return nil, errors.Wrap(err, "could not fetch gitlab composer projects")
	}

	packages := make(map[string]composer.PackageInfo)

	for _, project := range projects {
		if s.config.IsVendorAllowed(project.Vendor) {
			packages[project.Name] = s.createComposerPackageInfo(project)
		}
	}

	composerRepository := composer.Repository{
		Packages:    packages,
		NotifyBatch: "/notify",
	}
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
