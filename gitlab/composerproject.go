package gitlab

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"github.com/pkg/errors"

	"github.com/xanzy/go-gitlab"
)

type ComposerProject struct {
	Name         string
	Project      *gitlab.Project
	Head         *gitlab.Commit
	Tags         []*gitlab.Tag
	ComposerJson map[string]interface{}
}

func (project *ComposerProject) GitUrl() string {
	url := project.Project.SSHURLToRepo

	if url == "" { // maybe this should be an option
		return project.Project.HTTPURLToRepo
	}

	return url
}

func (project *ComposerProject) Type() string {
	if composerType, ok := project.ComposerJson["type"]; ok {
		return composerType.(string)
	}

	return "library" // because this is the default
}

func (c *Client) createComposerProject(project *gitlab.Project, file *gitlab.File) (*ComposerProject, error) {
	// determine composer project name and json file
	data, err := base64.StdEncoding.DecodeString(file.Content)
	if err != nil {
		return nil, err
	}

	var composerJson map[string]interface{}
	err = json.Unmarshal(data, &composerJson)
	if err != nil {
		return nil, errors.Wrapf(err, "could not parse composer.json in project %s", project.PathWithNamespace)
	}

	if _, ok := composerJson["name"]; !ok {
		return nil, fmt.Errorf("composer.json has no name in project %s", project.PathWithNamespace)
	}

	name := composerJson["name"].(string)

	// determine head commit
	commits, _, err := c.gitlab.Commits.ListCommits(project.ID, &gitlab.ListCommitsOptions{
		ListOptions: gitlab.ListOptions{
			Page:    0,
			PerPage: 1,
		},
	})
	if err != nil {
		return nil, errors.Wrapf(err, "could not determine HEAD commit in project %s", project.PathWithNamespace)
	}

	if len(commits) == 0 {
		return nil, fmt.Errorf("could not find any commits in project %s", project.PathWithNamespace)
	}

	headCommit := commits[0]

	// determine tags
	tags, _, err := c.gitlab.Tags.ListTags(project.ID, &gitlab.ListTagsOptions{})
	if err != nil {
		return nil, errors.Wrapf(err, "could not read tags in project %s", project.PathWithNamespace)
	}

	composerProject := ComposerProject{
		Name:         name,
		Project:      project,
		Head:         headCommit,
		Tags:         tags,
		ComposerJson: composerJson,
	}

	return &composerProject, nil
}
