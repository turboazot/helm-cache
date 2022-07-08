package entities

import (
	"fmt"

	"github.com/turboazot/helm-cache/pkg/utils"
)

type HelmReleaseFile struct {
	Name string `json:"name"`
	Data string `json:"data"`
}

func (f *HelmReleaseFile) Save(directory string) error {
	return utils.WriteStringToFile(fmt.Sprintf("%s/%s", directory, f.Name), f.Data)
}
