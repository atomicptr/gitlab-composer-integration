package service

import (
	"os"
	"path"
	"time"

	"github.com/patrickmn/go-cache"
	"github.com/pkg/errors"
)

const cacheKey = "gitlab-composer-packages-json-cache"

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
