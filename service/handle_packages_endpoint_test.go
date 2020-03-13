package service

import (
	"testing"

	"github.com/stretchr/testify/assert"
	goGitlab "github.com/xanzy/go-gitlab"

	"github.com/atomicptr/gitlab-composer-integration/gitlab"
)

func TestCreateHash(t *testing.T) {
	values := map[string]string{
		"":                                      "e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855",
		"test":                                  "9f86d081884c7d659a2feaa0c55ad015a3bf4f1b2b0b822cd15d6c15b0f00a08",
		"atomicptr/gitlab-composer-integration": "e2c9c61e77d12d7ba01249af09bfd08e9041d9e551134c372e6d74c4fad05164",
	}

	for value, expected := range values {
		actual, err := createHash([]byte(value))

		assert.Nil(t, err)
		assert.EqualValues(t, expected, actual)
	}
}

func TestGetProjectCacheIdentifier(t *testing.T) {
	assert.EqualValues(t, "project:atomicptr/package", getProjectCacheIdentifier("atomicptr/package"))
}

func TestGetProjectHashIdentifier(t *testing.T) {
	assert.EqualValues(
		t,
		"hash:e2c9c61e77d12d7ba01249af09bfd08e9041d9e551134c372e6d74c4fad05164",
		getProjectHashIdentifier("e2c9c61e77d12d7ba01249af09bfd08e9041d9e551134c372e6d74c4fad05164"),
	)
}

func TestCreateComposerPackageInfo(t *testing.T) {
	commit := goGitlab.Commit{
		ID: "9f86d081884c7d659a2feaa0c55ad015a3bf4f1b2b0b822cd15d6c15b0f00a08",
	}
	gitlabProject := goGitlab.Project{
		SSHURLToRepo: "ssh://git@gitlab.com:atomicptr/project.git",
	}
	project := gitlab.ComposerProject{
		Name:    "atomicptr/test-project",
		Head:    &commit,
		Project: &gitlabProject,
		Tags: []*goGitlab.Tag{
			&goGitlab.Tag{
				Name:   "v1.0.0",
				Commit: &commit,
			},
		},
	}

	packageCounter = 41
	packageInfo := createComposerPackageInfo(&project)

	assert.NotNil(t, packageInfo["dev-master"])
	assert.NotNil(t, packageInfo["v1.0.0"])
	assert.EqualValues(t, packageInfo["dev-master"].Uid, 42)
	assert.EqualValues(t, packageInfo["v1.0.0"].Uid, 43)
	assert.True(t, len(packageInfo) >= 2)

	for version, info := range packageInfo {
		assert.EqualValues(t, version, info.Version)
		assert.EqualValues(t, project.Name, info.Name)
		assert.EqualValues(t, project.Head.ID, info.Source.Reference)
		assert.EqualValues(t, project.GitUrl(), info.Source.Url)
		assert.EqualValues(t, project.Type(), info.Type)
	}
}

func TestNextPackageId(t *testing.T) {
	initialValue := int64(10)
	packageCounter = initialValue

	next := nextPackageId()

	assert.NotEqual(t, initialValue, next)
	assert.NotEqual(t, next, nextPackageId())
}
