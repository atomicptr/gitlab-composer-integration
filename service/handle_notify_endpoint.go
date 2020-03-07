package service

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
)

func (s *Service) handleNotifyEndpoint(writer http.ResponseWriter, request *http.Request) {
	if request.Method != "POST" {
		http.Error(writer, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
		return
	}

	writer.WriteHeader(http.StatusOK)

	data, err := ioutil.ReadAll(request.Body)
	if err != nil {
		s.logger.Println("err:", err)
		return
	}

	var download struct {
		Downloads []struct {
			Name    string `json:"name"`
			Version string `json:"version"`
		} `json:"downloads"`
	}

	err = json.Unmarshal(data, &download)
	if err != nil {
		s.logger.Println("err:", err)
		return
	}

	packages := ""
	for _, pkg := range download.Downloads {
		packages = fmt.Sprintf("%s\n\tPackage: %s, Version: %s", packages, pkg.Name, pkg.Version)
	}

	s.logger.Printf("Download from %s (%s)%s", request.UserAgent(), request.RemoteAddr, packages)
}
