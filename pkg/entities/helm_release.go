package entities

type HelmRelease struct {
	Name  string `json:"name" yaml:"name"`
	Chart struct {
		Metadata struct {
			ApiVersion   string                   `json:"apiVersion" yaml:"apiVersion"`
			Name         string                   `json:"name" yaml:"name"`
			Version      string                   `json:"version" yaml:"version"`
			KubeVersion  string                   `json:"kubeVersion,omitempty" yaml:"kubeVersion,omitempty"`
			Description  string                   `json:"description,omitempty" yaml:"description,omitempty"`
			Type         string                   `json:"type,omitempty" yaml:"type,omitempty"`
			Keywords     []string                 `json:"keywords,omitempty" yaml:"keywords,omitempty"`
			Home         string                   `json:"home,omitempty" yaml:"home,omitempty"`
			Sources      []string                 `json:"sources,omitempty" yaml:"sources,omitempty"`
			Dependencies []map[string]interface{} `json:"dependencies,omitempty" yaml:"dependencies,omitempty"`
			Maintainers  []map[string]string      `json:"maintainers,omitempty" yaml:"maintainers,omitempty"`
			Icon         string                   `json:"icon,omitempty" yaml:"icon,omitempty"`
			AppVersion   string                   `json:"appVersion,omitempty" yaml:"appVersion,omitempty"`
			Deprecated   bool                     `json:"deprecated,omitempty" yaml:"deprecated,omitempty"`
			Annotations  map[string]string        `json:"annotations,omitempty" yaml:"annotations,omitempty"`
		}
		Templates []HelmReleaseFile      `json:"templates" yaml:"templates"`
		Values    map[string]interface{} `json:"values" yaml:"values"`
		Files     []HelmReleaseFile      `json:"files" yaml:"files"`
	} `json:"chart" yaml:"chart"`
	IsSaved    bool
	IsPackaged bool
}
