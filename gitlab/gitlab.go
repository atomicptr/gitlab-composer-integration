package gitlab

import (
	"crypto/tls"
	"github.com/pkg/errors"
	"log"
	"net"
	"net/http"
	"time"

	"github.com/xanzy/go-gitlab"
)

const ApiSuffix = "/api/v4"

type Client struct {
	gitlab *gitlab.Client
	logger *log.Logger
}

func New(baseUrl, token string, logger *log.Logger) *Client {
	httpClient := &http.Client{
		Transport: &http.Transport{
			Proxy: http.ProxyFromEnvironment,
			DialContext: (&net.Dialer{
				Timeout:   30 * time.Second,
				KeepAlive: 30 * time.Second,
			}).DialContext,
			ForceAttemptHTTP2:     true,
			IdleConnTimeout:       90 * time.Second,
			TLSHandshakeTimeout:   10 * time.Second,
			ExpectContinueTimeout: 1 * time.Second,
			TLSClientConfig:       &tls.Config{InsecureSkipVerify: true},
		},
	}

	git := gitlab.NewClient(httpClient, token)
	git.UserAgent = "gitlab-composer-integration"

	gitlabBaseUrl := baseUrl + ApiSuffix
	logger.Printf("using \"%s\" as gitlab base url", gitlabBaseUrl)
	err := git.SetBaseURL(gitlabBaseUrl)
	if err != nil {
		logger.Println(err)
		return nil
	}

	client := &Client{
		gitlab: git,
		logger: logger,
	}
	return client
}

func (c *Client) Validate() error {
	_, _, err := c.gitlab.Projects.ListProjects(&gitlab.ListProjectsOptions{})
	if err != nil {
		return err
	}
	return nil
}

func (c *Client) FindAllComposerProjects() ([]*ComposerProject, error) {
	running := true

	const ComposerFileName = "composer.json"
	const PageSize = 50

	var composerProjects []*ComposerProject

	page := 0
	for running {
		projects, _, err := c.gitlab.Projects.ListProjects(&gitlab.ListProjectsOptions{
			ListOptions: gitlab.ListOptions{
				Page:    page,
				PerPage: PageSize,
			},
		})
		if err != nil {
			return nil, err
		}

		for _, project := range projects {
			file, _, err := c.gitlab.RepositoryFiles.GetFile(project.ID, ComposerFileName, &gitlab.GetFileOptions{
				Ref: gitlab.String(project.DefaultBranch),
			})

			if err == nil && file != nil {
				composerProject, err := c.createComposerProject(project, file)
				if err != nil {
					c.logger.Println(errors.Wrap(err, "error: invalid composer project"))
					continue
				}
				composerProjects = append(composerProjects, composerProject)
			}
		}

		page++
		if len(projects) < PageSize {
			running = false
		}
	}

	c.logger.Printf("%d projects found", len(composerProjects))
	return composerProjects, nil
}
