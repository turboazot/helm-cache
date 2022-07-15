package entities

import (
	"helm.sh/helm/v3/pkg/release"
)

type HelmRelease struct {
	Release    *release.Release
	IsSaved    bool
	IsPackaged bool
}
