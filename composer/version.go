package composer

type VersionInfo struct {
	Name    string     `json:"name"`
	Source  SourceInfo `json:"source"`
	Type    string     `json:"type"`
	Version string     `json:"version"`
}
