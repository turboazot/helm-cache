package entities

import (
	"encoding/base64"
	"fmt"

	"github.com/turboazot/helm-cache/pkg/utils"
)

type HelmReleaseFile struct {
	Name string `json:"name"`
	Data string `json:"data"`
}

func (f *HelmReleaseFile) GetData() (string, error) {
	b, err := base64.StdEncoding.DecodeString(f.Data)
	return string(b), err
}

func (f *HelmReleaseFile) Save(directory string) error {
	data, err := f.GetData()
	if err != nil {
		return err
	}
	return utils.WriteStringToFile(fmt.Sprintf("%s/%s", directory, f.Name), data)
}
