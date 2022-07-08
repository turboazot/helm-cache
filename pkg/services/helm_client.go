package services

import (
	"bytes"
	"compress/gzip"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"strings"

	"github.com/turboazot/helm-cache/pkg/entities"
	"github.com/turboazot/helm-cache/pkg/utils"
	"helm.sh/helm/v3/pkg/action"
	"helm.sh/helm/v3/pkg/cli"
	"helm.sh/helm/v3/pkg/downloader"
	"helm.sh/helm/v3/pkg/getter"
	v1 "k8s.io/api/core/v1"

	"go.uber.org/zap"
)

func saveHelmReleaseFileCollection(directory string, files *[]entities.HelmReleaseFile) error {
	for _, f := range *files {
		err := f.Save(directory)
		if err != nil {
			return err
		}
	}

	return nil
}

type HelmClient struct {
	ActionConfig            *action.Configuration
	Settings                *cli.EnvSettings
	RawChartsDirectory      string
	PackagedChartsDirectory string
}

func NewHelmClient(homeDirectory string) (*HelmClient, error) {
	rawChartsDirectory := fmt.Sprintf("%s/data/raw", homeDirectory)
	if err := os.MkdirAll(rawChartsDirectory, 0755); err != nil {
		return nil, err
	}
	packagedChartsDirectory := fmt.Sprintf("%s/data/packaged", homeDirectory)
	if err := os.MkdirAll(packagedChartsDirectory, 0755); err != nil {
		return nil, err
	}

	return &HelmClient{
		ActionConfig:            new(action.Configuration),
		Settings:                cli.New(),
		RawChartsDirectory:      rawChartsDirectory,
		PackagedChartsDirectory: packagedChartsDirectory,
	}, nil
}

func (c *HelmClient) GetHelmRelease(s *entities.HelmReleaseSecret) (*entities.HelmRelease, error) {
	var r entities.HelmRelease
	r.IsSaved = false

	if _, releaseKeyExists := s.Secret.Data["release"]; !releaseKeyExists {
		return nil, errors.New(fmt.Sprintf("Release secret %s doesn't contain release key in data", s.Secret.Name))
	}

	var base64DecodedBytes []byte

	base64DecodedBytes, err := base64.StdEncoding.DecodeString(string(s.Secret.Data["release"]))
	if err != nil {
		return nil, err
	}

	g, err := gzip.NewReader(bytes.NewReader(base64DecodedBytes))
	if err != nil {
		return nil, err
	}

	decodedBytes, err := ioutil.ReadAll(g)
	if err != nil {
		return nil, err
	}

	err = json.Unmarshal(decodedBytes, &r)
	if err != nil {
		return nil, err
	}

	_, err = os.Stat(fmt.Sprintf("%s/%s-%s", c.RawChartsDirectory, r.Chart.Metadata.Name, r.Chart.Metadata.Version))
	if err == nil {
		r.IsSaved = true
	}

	_, err = os.Stat(fmt.Sprintf("%s/%s-%s.tgz", c.PackagedChartsDirectory, r.Chart.Metadata.Name, r.Chart.Metadata.Version))
	if err == nil {
		r.IsPackaged = true
	}

	return &r, nil
}

func (c *HelmClient) GetLastRevisionReleaseSecretsMap(secrets *v1.SecretList) (map[string]*entities.HelmReleaseSecret, error) {
	result := make(map[string]*entities.HelmReleaseSecret)
	releaseIdLastRevisionMap := make(map[string]int)
	for index, secret := range secrets.Items {
		if !strings.HasPrefix(secret.Name, "sh.helm.release.v1") {
			continue
		}

		rs := entities.NewHelmReleaseSecret(&secrets.Items[index])
		releaseName, releaseRevision, err := rs.GetReleaseNameAndRevision()
		if err != nil {
			return nil, err
		}
		releaseNamespace := rs.Secret.Namespace
		releaseID := fmt.Sprintf("%s-%s", releaseNamespace, releaseName)
		if _, releaseExists := releaseIdLastRevisionMap[releaseID]; !releaseExists || releaseIdLastRevisionMap[releaseID] < releaseRevision {
			releaseIdLastRevisionMap[releaseID] = releaseRevision
			result[releaseID] = rs
		}
	}

	return result, nil
}

func (c *HelmClient) SaveRawChart(r *entities.HelmRelease) error {
	if r.IsSaved {
		zap.L().Sugar().Infof("Chart %s-%s already saved in local filesystem", r.Chart.Metadata.Name, r.Chart.Metadata.Version)
		return nil
	}
	directory := fmt.Sprintf("%s/%s-%s", c.RawChartsDirectory, r.Chart.Metadata.Name, r.Chart.Metadata.Version)

	err := utils.WriteYamlToFile(&r.Chart.Values, fmt.Sprintf("%s/%s", directory, "values.yaml"))
	if err != nil {
		return err
	}
	err = utils.WriteYamlToFile(&r.Chart.Metadata, fmt.Sprintf("%s/%s", directory, "Chart.yaml"))
	if err != nil {
		return err
	}
	err = saveHelmReleaseFileCollection(fmt.Sprintf("%s/templates", directory), &r.Chart.Templates)
	if err != nil {
		return err
	}

	err = saveHelmReleaseFileCollection(directory, &r.Chart.Files)
	if err != nil {
		return err
	}

	r.IsSaved = true

	zap.L().Sugar().Infof("Successfully saved raw chart: %s-%s", r.Chart.Metadata.Name, r.Chart.Metadata.Version)

	return nil
}

func (c *HelmClient) GetReleasePackageFile(r *entities.HelmRelease) (*os.File, error) {
	if !r.IsPackaged {
		return nil, errors.New(fmt.Sprintf("Release %s hasn't been saved yet", r.Name))
	}

	return os.Open(fmt.Sprintf("%s/%s-%s.tgz", c.PackagedChartsDirectory, r.Chart.Metadata.Name, r.Chart.Metadata.Version))
}

func (c *HelmClient) Package(chartName string, chartVersion string) error {
	path := fmt.Sprintf("%s/%s-%s", c.RawChartsDirectory, chartName, chartVersion)

	client := action.NewPackage()

	w := zap.NewStdLog(zap.L()).Writer()

	client.RepositoryConfig = c.Settings.RepositoryConfig
	client.RepositoryCache = c.Settings.RepositoryCache
	client.Destination = c.PackagedChartsDirectory

	downloadManager := &downloader.Manager{
		Out:              w,
		ChartPath:        path,
		Keyring:          client.Keyring,
		SkipUpdate:       false,
		Getters:          getter.All(c.Settings),
		RegistryClient:   c.ActionConfig.RegistryClient,
		RepositoryConfig: c.Settings.RepositoryConfig,
		RepositoryCache:  c.Settings.RepositoryCache,
		Debug:            c.Settings.Debug,
	}

	if err := downloadManager.Update(); err != nil {
		return err
	}

	p, err := client.Run(path, make(map[string]interface{}))
	if err != nil {
		return err
	}

	zap.L().Sugar().Infof("Successfully packaged chart and saved it to: %s", p)
	return nil
}
