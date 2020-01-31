package composer

import "encoding/json"

type PackageInfo map[string]VersionInfo

type Repository struct {
	Packages map[string]PackageInfo `json:"packages"`
}

func Example() Repository {
	return Repository{
		Packages: map[string]PackageInfo{
			"test/test": {
				"dev-master": {
					Name: "test/test",
					Source: SourceInfo{
						Reference: "882816c7c05b5b5704e84bdb0f7ad69230df3c0c",
						Type:      "git",
						Url:       "git@git.domain.com:test/test.git",
					},
					Type:    "project",
					Version: "dev-master",
				},
				"v1.5": {
					Name: "test/test",
					Source: SourceInfo{
						Reference: "882816c7c05b5b5704e84bdb0f7ad69230df3c0c",
						Type:      "git",
						Url:       "git@git.domain.com:test/test.git",
					},
					Type:    "project",
					Version: "v1.5",
				},
			},
		},
	}
}

func (r *Repository) ToJson() ([]byte, error) {
	return json.Marshal(r)
}
