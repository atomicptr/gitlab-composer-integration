package gitlab

import (
	"fmt"
	"net/http"
	"net/http/httptest"

	"github.com/xanzy/go-gitlab"
)

func gitlabTestServerSetup() (*http.ServeMux, *httptest.Server, *gitlab.Client) {
	// mux is the HTTP request multiplexer used with the test server.
	mux := http.NewServeMux()

	// server is a test HTTP server used to provide mock API responses.
	server := httptest.NewServer(mux)

	// client is the Gitlab client being tested.
	client := gitlab.NewClient(nil, "")
	_ = client.SetBaseURL(server.URL)

	return mux, server, client
}

func registerApiResult(mux *http.ServeMux, path, result string) {
	mux.HandleFunc(fmt.Sprintf("%s/%s", ApiSuffix, path), func(writer http.ResponseWriter, _ *http.Request) {
		_, _ = fmt.Fprint(writer, result)
	})
}
