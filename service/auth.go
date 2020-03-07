package service

import (
	"bytes"
	"encoding/base64"
	"fmt"
	"net/http"
	"strings"
)

const authRealm = "Composer Repository"

func basicAuth(username, password string, handlerFunc func(http.ResponseWriter, *http.Request)) func(http.ResponseWriter, *http.Request) {
	if username == "" || password == "" {
		return handlerFunc
	}

	return func(writer http.ResponseWriter, request *http.Request) {
		if !isAuthenticated(username, password, request) {
			requestAuthentication(writer)
			return
		}
		handlerFunc(writer, request)
	}
}

func isAuthenticated(username, password string, request *http.Request) bool {
	basicPrefix := "Basic "
	auth := request.Header.Get("Authorization")

	// confirm the user is sending basic authentication
	if !strings.HasPrefix(auth, basicPrefix) {
		return false
	}

	// skip "Basic "
	base64CredentialsString := auth[len(basicPrefix):]

	// decode base64 string
	decodedCredentialsString, err := base64.StdEncoding.DecodeString(base64CredentialsString)
	if err != nil {
		return false
	}

	// split on the first ":" character, any colons after that are part of the password
	credentialParts := bytes.SplitN(decodedCredentialsString, []byte(":"), 2)

	if len(credentialParts) != 2 {
		return false
	}

	reqUsername := string(credentialParts[0])
	reqPassword := string(credentialParts[1])

	return username == reqUsername && password == reqPassword
}

func requestAuthentication(writer http.ResponseWriter) {
	writer.Header().Set("WWW-Authenticate", fmt.Sprintf("Basic realm=\"%s\", charset=\"UTF-8\"", authRealm))
	http.Error(writer, http.StatusText(http.StatusUnauthorized), http.StatusUnauthorized)
}
