package gitlab

import (
	"encoding/base64"
	"fmt"
	"github.com/stretchr/testify/assert"
	"io/ioutil"
	"log"
	"testing"
)

func TestNewWithInvalidBaseUrl(t *testing.T) {
	client := New("https://invalid url", "this is ma token", log.New(ioutil.Discard, "", 0))
	assert.Nil(t, client)
}

func TestNew(t *testing.T) {
	client := New("https://gitlab.com", "this is ma token", log.New(ioutil.Discard, "", 0))
	assert.EqualValues(t, "https://gitlab.com/api/v4/", client.gitlab.BaseURL().String())
}

func TestValidateApiError(t *testing.T) {
	_, _, gitlabClient := gitlabTestServerSetup()

	client := Client{
		gitlab: gitlabClient,
		logger: log.New(ioutil.Discard, "", 0),
	}

	assert.NotNil(t, client.Validate())
}

func TestValidate(t *testing.T) {
	mux, _, gitlabClient := gitlabTestServerSetup()

	registerApiResult(mux, "projects", "[]")

	client := Client{
		gitlab: gitlabClient,
		logger: log.New(ioutil.Discard, "", 0),
	}

	assert.Nil(t, client.Validate())
}

func TestFindAllComposerProjectsListApiError(t *testing.T) {
	_, _, gitlabClient := gitlabTestServerSetup()

	client := Client{
		gitlab: gitlabClient,
		logger: log.New(ioutil.Discard, "", 0),
	}

	_, err := client.FindAllComposerProjects()

	assert.NotNil(t, err)
}

func TestFindAllComposerProjectsInvalidJson(t *testing.T) {
	mux, _, gitlabClient := gitlabTestServerSetup()

	const composerJson = `{
		"test": 42
	}`

	fileData := fmt.Sprintf(`{
			"file_name": "composer.json",
			"file_path": "composer.json",
			"size": %d,
			"encoding": "base64",
			"content": "%s",
			"content_sha256": "...",
			"ref": "master",
			"blob_id": "79f7bbd25901e8334750839545a9bd021f0e4c83",
			"commit_id": "d5a3ff139356ce33e37e73add446f16869741b50",
			"last_commit_id": "570e7b2abdd848b95f2f578043fc23bd6f6fd24d"
		}`,
		12,
		base64.StdEncoding.EncodeToString([]byte(composerJson)),
	)

	registerApiResult(mux, "projects", `[{"id": 0}]`)
	registerApiResult(mux, "projects/0/repository/commits", `[{"id": "1234"}]`)
	registerApiResult(mux, "projects/0/repository/tags", `[{"name": "v1.0.0"}]`)
	registerApiResult(
		mux,
		"projects/0/repository/files/composer.json",
		fileData,
	)

	client := Client{
		gitlab: gitlabClient,
		logger: log.New(ioutil.Discard, "", 0),
	}

	_, err := client.FindAllComposerProjects()

	assert.Nil(t, err)
}

func TestFindAllComposerProjects(t *testing.T) {
	mux, _, gitlabClient := gitlabTestServerSetup()

	const composerJson = `{
		"name": "atomicptr/test-package"
	}`

	fileData := fmt.Sprintf(`{
			"file_name": "composer.json",
			"file_path": "composer.json",
			"size": %d,
			"encoding": "base64",
			"content": "%s",
			"content_sha256": "...",
			"ref": "master",
			"blob_id": "79f7bbd25901e8334750839545a9bd021f0e4c83",
			"commit_id": "d5a3ff139356ce33e37e73add446f16869741b50",
			"last_commit_id": "570e7b2abdd848b95f2f578043fc23bd6f6fd24d"
		}`,
		12,
		base64.StdEncoding.EncodeToString([]byte(composerJson)),
	)

	registerApiResult(mux, "projects", `[{"id": 0}]`)
	registerApiResult(mux, "projects/0/repository/commits", `[{"id": "1234"}]`)
	registerApiResult(mux, "projects/0/repository/tags", `[{"name": "v1.0.0"}]`)
	registerApiResult(
		mux,
		"projects/0/repository/files/composer.json",
		fileData,
	)

	client := Client{
		gitlab: gitlabClient,
		logger: log.New(ioutil.Discard, "", 0),
	}

	_, err := client.FindAllComposerProjects()

	assert.Nil(t, err)
}
