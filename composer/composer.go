package composer

import "encoding/json"

type PackageInfo map[string]VersionInfo

type Repository struct {
	Packages    map[string]PackageInfo `json:"packages"`
	NotifyBatch string                 `json:"notify-batch"`
}

func (r *Repository) ToJson() ([]byte, error) {
	return json.Marshal(r)
}
