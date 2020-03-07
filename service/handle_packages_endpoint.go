package service

import (
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/patrickmn/go-cache"
	"github.com/pkg/errors"

	"github.com/atomicptr/gitlab-composer-integration/composer"
	"github.com/atomicptr/gitlab-composer-integration/gitlab"
)

var packageCounter int64

func (s *Service) handlePackagesJsonEndpoint(writer http.ResponseWriter, request *http.Request) {
	s.logger.Printf("Request to \"%s\" from %s (%s)\n", request.URL, request.UserAgent(), request.RemoteAddr)

	writer.Header().Set("Content-Type", "application/json")

	// busy loop until you can get a cache...
	for {
		if content, found := s.cache.Get(indexCacheKey); found {
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

	packageCounter = 0

	providers := make(map[string]composer.Provider)

	for _, project := range projects {
		if s.config.IsVendorAllowed(project.Vendor) {
			packages := map[string]composer.PackageInfo{}
			packages[project.Name] = createComposerPackageInfo(project)

			packageData := composer.ProviderRepository{
				Packages: packages,
			}

			data, err := json.Marshal(packageData)
			if err != nil {
				s.logger.Println(errors.Wrapf(err, "could not cache project: %s", project.Name))
				continue
			}

			hash, err := createHash(data)
			if err != nil {
				s.logger.Println(errors.Wrap(err, "could not create sha256 hash"))
				continue
			}

			// store package in cache
			s.cache.Set(
				getProjectCacheIdentifier(project.Name),
				data,
				cache.DefaultExpiration,
			)

			// store hash
			s.cache.Set(
				getProjectHashIdentifier(project.Name),
				hash,
				cache.DefaultExpiration,
			)

			providers[project.Name] = composer.Provider{Sha256: hash}
		}
	}

	composerRepository := composer.Repository{
		Packages:     []struct{}{},
		NotifyBatch:  "/notify",
		ProvidersUrl: "/p?package=%package%&hash=%hash%",
		Providers:    providers,
	}
	return &composerRepository, nil
}

func createHash(str []byte) (string, error) {
	hasher := sha256.New()
	_, err := hasher.Write(str)
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("%x", hasher.Sum(nil)), nil
}

func getProjectCacheIdentifier(projectName string) string {
	return fmt.Sprintf("project:%s", projectName)
}

func getProjectHashIdentifier(projectName string) string {
	return fmt.Sprintf("hash:%s", projectName)
}

func createComposerPackageInfo(project *gitlab.ComposerProject) composer.PackageInfo {
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
		Uid:     nextPackageId(),
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
			Uid:     nextPackageId(),
		}
	}

	return packageInfo
}

func nextPackageId() int64 {
	packageCounter++
	return packageCounter
}
