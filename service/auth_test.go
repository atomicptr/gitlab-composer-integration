package service

import (
	"fmt"
	"github.com/stretchr/testify/assert"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

// base64 of "username:pass:word"
const authString = "Basic dXNlcm5hbWU6cGFzczp3b3Jk"

func makeTestRequest(auth string) *http.Request {
	request := httptest.NewRequest("GET", "/", strings.NewReader(""))
	request.Header.Set("Authorization", auth)
	return request
}

func TestBasicAuthUnauthorized(t *testing.T) {
	handlerFunc := func(writer http.ResponseWriter, request *http.Request) {}
	wrappedHandlerFunc := basicAuth("username", "pass:word", handlerFunc)
	server := httptest.NewServer(http.HandlerFunc(wrappedHandlerFunc))
	defer server.Close()

	resp, err := http.Get(server.URL)
	assert.Nil(t, err)

	assert.EqualValues(t, http.StatusUnauthorized, resp.StatusCode)
}

func TestBasicAuthWithoutCredentials(t *testing.T) {
	handlerFunc := func(writer http.ResponseWriter, request *http.Request) {
		writer.WriteHeader(http.StatusOK)
		_, err := writer.Write([]byte("OK"))
		assert.Nil(t, err)
	}
	wrappedHandlerFunc := basicAuth("", "", handlerFunc)
	server := httptest.NewServer(http.HandlerFunc(wrappedHandlerFunc))
	defer server.Close()

	request, err := http.NewRequest("GET", server.URL, nil)
	assert.Nil(t, err)

	response, err := http.DefaultClient.Do(request)
	assert.Nil(t, err)

	assert.EqualValues(t, http.StatusOK, response.StatusCode)

	body, err := ioutil.ReadAll(response.Body)
	assert.Nil(t, err)

	assert.EqualValues(t, "OK", string(body))
}

func TestBasicAuth(t *testing.T) {
	handlerFunc := func(writer http.ResponseWriter, request *http.Request) {
		writer.WriteHeader(http.StatusOK)
		_, err := writer.Write([]byte("OK"))
		assert.Nil(t, err)
	}
	wrappedHandlerFunc := basicAuth("username", "pass:word", handlerFunc)
	server := httptest.NewServer(http.HandlerFunc(wrappedHandlerFunc))
	defer server.Close()

	request, err := http.NewRequest("GET", server.URL, nil)
	assert.Nil(t, err)

	request.SetBasicAuth("username", "pass:word")

	response, err := http.DefaultClient.Do(request)
	assert.Nil(t, err)

	assert.EqualValues(t, http.StatusOK, response.StatusCode)

	body, err := ioutil.ReadAll(response.Body)
	assert.Nil(t, err)

	assert.EqualValues(t, "OK", string(body))
}

func TestIsAuthenticatedInvalidPrefix(t *testing.T) {
	customAuthString := "Secret dXNlcm5hbWU6cGFzczp3b3Jk"
	assert.False(t, isAuthenticated("username", "pass:word", makeTestRequest(customAuthString)))
}

func TestIsAuthenticatedInvalidBase64String(t *testing.T) {
	customAuthString := "Basic This is not a base64 string"
	assert.False(t, isAuthenticated("username", "pass:word", makeTestRequest(customAuthString)))
}

func TestIsAuthenticatedInvalidAuthStringCreds(t *testing.T) {
	customAuthString := "Basic dXNlcm5hbWVwYXNzd29yZA=="
	assert.False(t, isAuthenticated("username", "pass:word", makeTestRequest(customAuthString)))
}

func TestIsAuthenticatedInvalidCredentials(t *testing.T) {
	assert.False(t, isAuthenticated("username", "password", makeTestRequest(authString)))
}

func TestIsAuthenticated(t *testing.T) {
	assert.True(t, isAuthenticated("username", "pass:word", makeTestRequest(authString)))
}

func TestRequestAuthentication(t *testing.T) {
	recorder := httptest.NewRecorder()
	requestAuthentication(recorder)

	authHeader := fmt.Sprintf("Basic realm=\"%s\", charset=\"UTF-8\"", authRealm)

	assert.EqualValues(t, http.StatusUnauthorized, recorder.Code)
	assert.EqualValues(t, authHeader, recorder.Header().Get("WWW-Authenticate"))
}
