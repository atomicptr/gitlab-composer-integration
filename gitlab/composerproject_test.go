package gitlab

import (
	"encoding/base64"
	"io/ioutil"
	"log"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/xanzy/go-gitlab"
)

func TestGirUrlWithSshUrl(t *testing.T) {
	sshUrl := "ssh://git@gitlab.com:test/project.git"
	project := ComposerProject{
		Project: &gitlab.Project{SSHURLToRepo: sshUrl},
	}

	assert.EqualValues(t, sshUrl, project.GitUrl())
}

func TestGitUrlWithHttpUrl(t *testing.T) {
	httpUrl := "https://gitlab.com/test/project.git"
	project := ComposerProject{
		Project: &gitlab.Project{HTTPURLToRepo: httpUrl},
	}

	assert.EqualValues(t, httpUrl, project.GitUrl())
}

func TestType(t *testing.T) {
	project := ComposerProject{
		ComposerJson: map[string]interface{}{
			"type": "website",
		},
	}

	assert.EqualValues(t, "website", project.Type())
}

func TestTypeDefault(t *testing.T) {
	project := ComposerProject{}

	assert.EqualValues(t, "library", project.Type())
}

func TestExtractVendorFromComposerName(t *testing.T) {
	values := map[string]string{
		"package-name-without-vendor": "",
		"vendor/package":              "vendor",
		"vendor/package/sub-package":  "vendor",
	}

	for value, expected := range values {
		assert.EqualValues(t, expected, extractVendorFromComposerName(value))
	}
}

func TestCreateComposerProjectInvalidContent(t *testing.T) {
	_, _, gitlabClient := gitlabTestServerSetup()
	_, err := tryCreateComposerProjectWithContent(gitlabClient, "this is not a base64 string! :)")
	assert.NotNil(t, err)
}

func TestCreateComposerProjectInvalidJson(t *testing.T) {
	_, _, gitlabClient := gitlabTestServerSetup()
	_, err := tryCreateComposerProjectWithContent(
		gitlabClient,
		base64.StdEncoding.EncodeToString(
			[]byte(`{"name": atomicpt...`),
		),
	)
	assert.NotNil(t, err)
}

func TestCreateComposerProjectInvalidJsonWithoutName(t *testing.T) {
	_, _, gitlabClient := gitlabTestServerSetup()
	_, err := tryCreateComposerProjectWithContent(
		gitlabClient,
		base64.StdEncoding.EncodeToString(
			[]byte(`{"field": 42}`),
		),
	)
	assert.NotNil(t, err)
}

func TestCreateComposerProjectCommitApiError(t *testing.T) {
	const composerJson = `{
		"name": "atomicptr/test-package"
	}`

	_, _, gitlabClient := gitlabTestServerSetup()

	_, err := tryCreateComposerProjectWithContent(
		gitlabClient,
		base64.StdEncoding.EncodeToString([]byte(composerJson)),
	)

	assert.NotNil(t, err)
}

func TestCreateComposerProjectNoHeadCommit(t *testing.T) {
	const composerJson = `{
		"name": "atomicptr/test-package"
	}`

	mux, _, gitlabClient := gitlabTestServerSetup()

	registerApiResult(mux, "projects/0/repository/commits", `[]`)

	_, err := tryCreateComposerProjectWithContent(
		gitlabClient,
		base64.StdEncoding.EncodeToString([]byte(composerJson)),
	)

	assert.NotNil(t, err)
}

func TestCreateComposerProjectTagsApiError(t *testing.T) {
	const composerJson = `{
		"name": "atomicptr/test-package"
	}`

	mux, _, gitlabClient := gitlabTestServerSetup()

	registerApiResult(mux, "projects/0/repository/commits", `[{"id": "1234"}]`)

	_, err := tryCreateComposerProjectWithContent(
		gitlabClient,
		base64.StdEncoding.EncodeToString([]byte(composerJson)),
	)

	assert.NotNil(t, err)
}

func TestCreateComposerProject(t *testing.T) {
	const composerJson = `{
		"name": "atomicptr/test-package"
	}`

	mux, _, gitlabClient := gitlabTestServerSetup()

	registerApiResult(mux, "projects/0/repository/commits", `[{"id": "1234"}]`)
	registerApiResult(mux, "projects/0/repository/tags", `[{"name": "v1.0.0"}]`)

	_, err := tryCreateComposerProjectWithContent(
		gitlabClient,
		base64.StdEncoding.EncodeToString([]byte(composerJson)),
	)

	assert.Nil(t, err)
}

func tryCreateComposerProjectWithContent(gitlabClient *gitlab.Client, content string) (*ComposerProject, error) {
	client := Client{
		gitlab: gitlabClient,
		logger: log.New(ioutil.Discard, "", 0),
	}

	return client.createComposerProject(
		&gitlab.Project{},
		&gitlab.File{
			Content: content,
		},
	)
}
