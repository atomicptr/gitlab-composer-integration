package service

import (
	"net/http"
)

func (s *Service) handleProviderEndpoint(writer http.ResponseWriter, request *http.Request) {
	s.logger.Printf("Request to \"%s\" from %s (%s)\n", request.URL, request.UserAgent(), request.RemoteAddr)

	writer.Header().Set("Content-Type", "application/json")

	query := request.URL.Query()

	packageName := query.Get("package")
	hash := query.Get("hash")

	hashData, ok := s.cache.Get(getProjectHashIdentifier(packageName))
	if !ok || hash != hashData {
		s.logger.Printf("could not find package %s (hash: %s)\n", packageName, hash)
		writer.WriteHeader(http.StatusNotFound)
		return
	}

	data, ok := s.cache.Get(getProjectCacheIdentifier(packageName))
	if !ok {
		s.logger.Printf("could not find package %s (hash %s)\n", packageName, hash)
		writer.WriteHeader(http.StatusNotFound)
		return
	}

	_, err := writer.Write(data.([]byte))
	if err != nil {
		s.logger.Println(err)
	}
}
