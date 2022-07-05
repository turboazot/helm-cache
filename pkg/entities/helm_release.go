package entities

type HelmRelease struct {
	Name  string `json:"name"`
	Chart struct {
		Metadata struct {
			ApiVersion   string                  `json:"apiVersion"`
			Name         string                  `json:"name"`
			Version      string                  `json:"version"`
			KubeVersion  string                  `json:"kubeVersion,omitempty"`
			Description  string                  `json:"description,omitempty"`
			Type         string                  `json:"type,omitempty"`
			Keywords     []string                `json:"keywords,omitempty"`
			Home         string                  `json:"home,omitempty"`
			Sources      []string                `json:"sources,omitempty"`
			Dependencies []HelmReleaseDependency `json:"dependencies,omitempty"`
			Maintainers  []map[string]string     `json:"maintainers,omitempty"`
			Icon         string                  `json:"icon,omitempty"`
			AppVersion   string                  `json:"appVersion,omitempty"`
			Deprecated   bool                    `json:"deprecated,omitempty"`
			Annotations  map[string]string       `json:"annotations,omitempty"`
		}
		Templates []HelmReleaseFile      `json:"templates"`
		Values    map[string]interface{} `json:"values"`
		Files     []HelmReleaseFile      `json:"files"`
	} `json:"chart"`
	IsSaved    bool
	IsPackaged bool
}
