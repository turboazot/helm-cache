package entities

import (
	"errors"
	"strconv"
	"strings"

	v1 "k8s.io/api/core/v1"
)

type HelmReleaseSecret struct {
	Secret *v1.Secret
}

func NewHelmReleaseSecret(secret *v1.Secret) *HelmReleaseSecret {
	return &HelmReleaseSecret{
		Secret: secret,
	}
}

func (s *HelmReleaseSecret) GetReleaseNameAndRevision() (string, int, error) {
	secretNameSplitted := strings.Split(s.Secret.Name, ".")
	if len(secretNameSplitted) < 6 {
		return "", 0, errors.New("This secret is not helm release secret")
	}
	secretReleaseName := secretNameSplitted[4]
	secretRevisionInt, err := strconv.Atoi(strings.Replace(secretNameSplitted[5], "v", "", -1))
	return secretReleaseName, secretRevisionInt, err
}
