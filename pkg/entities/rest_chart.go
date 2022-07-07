package entities

type RestChart struct {
	Name        string              `json:"name"`
	Home        string              `json:"home"`
	Sources     []string            `json:"sources,omitempty"`
	Version     string              `json:"version"`
	Description string              `json:"description"`
	Keywords    []string            `json:"keywords,omitempty"`
	Maintainers []map[string]string `json:"maintainers,omitempty"`
	Icon        string              `json:"icon"`
	ApiVersion  string              `json:"apiVersion"`
	AppVersion  string              `json:"appVersion"`
	Urls        []string            `json:"urls"`
	Created     string              `json:"created"`
	Digest      string              `json:"digest"`
}
