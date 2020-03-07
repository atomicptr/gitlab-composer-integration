package composer

import "encoding/json"

type PackageInfo map[string]VersionInfo

type Repository struct {
	Packages     []struct{}          `json:"packages"`
	NotifyBatch  string              `json:"notify-batch"`
	ProvidersUrl string              `json:"providers-url"`
	Providers    map[string]Provider `json:"providers"`
}

type Provider struct {
	Sha256 string `json:"sha256"`
}

type ProviderRepository struct {
	Packages map[string]PackageInfo `json:"packages"`
}

func (r *Repository) ToJson() ([]byte, error) {
	return json.Marshal(r)
}
