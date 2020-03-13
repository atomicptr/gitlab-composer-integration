package gitlab

import (
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
