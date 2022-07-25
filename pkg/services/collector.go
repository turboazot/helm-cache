package services

import (
	"context"

	"go.uber.org/zap"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
)

type Collector struct {
	HelmClient          *HelmClient
	ChartmuseumClient   *ChartmuseumClient
	KubernetesClientset *kubernetes.Clientset
}

func NewCollector(helmClient *HelmClient, chartmuseumClient *ChartmuseumClient, kubeconfigPath string) (*Collector, error) {
	var config *rest.Config
	var err error

	if kubeconfigPath == "" {
		zap.L().Sugar().Info("Using in-cluster kubeconfig")
		config, err = rest.InClusterConfig()
		if err != nil {
			return nil, err
		}
	} else {
		zap.L().Sugar().Infof("Using %s kubeconfig", kubeconfigPath)
		config, err = clientcmd.BuildConfigFromFlags("", kubeconfigPath)
		if err != nil {
			return nil, err
		}
	}

	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		return nil, err
	}

	return &Collector{
		HelmClient:          helmClient,
		ChartmuseumClient:   chartmuseumClient,
		KubernetesClientset: clientset,
	}, nil
}

func (c *Collector) CheckAllSecrets() error {
	secrets, err := c.KubernetesClientset.CoreV1().Secrets("").List(context.TODO(), metav1.ListOptions{})
	if err != nil {
		return err
	}

	rsMap, err := c.HelmClient.GetLastRevisionReleaseSecretsMap(secrets)
	if err != nil {
		return err
	}

	for _, rs := range rsMap {
		zap.L().Sugar().Infof("Checking secret %s...", rs.Secret.Name)

		r, err := c.HelmClient.GetHelmRelease(rs)
		if err != nil {
			zap.L().Sugar().Infof("Can't decode release from secret %s: %v", rs.Secret.Name, err)
			continue
		}

		if c.ChartmuseumClient.IsActive() && c.ChartmuseumClient.IsExists(r.Release.Chart.Metadata.Name, r.Release.Chart.Metadata.Version) {
			zap.L().Sugar().Infof("Chart %s-%s already exists in the chartmuseum", r.Release.Chart.Metadata.Name, r.Release.Chart.Metadata.Version)
			continue
		}

		if err := c.HelmClient.SaveRawChart(r); err != nil {
			zap.L().Sugar().Infof("Can't save %s-%s chart in local filesystem: %v", r.Release.Chart.Metadata.Name, r.Release.Chart.Metadata.Version, err)
			continue
		}

		if r.IsPackaged {
			zap.L().Sugar().Infof("Chart %s-%s is already packaged in local filesystem", r.Release.Chart.Metadata.Name, r.Release.Chart.Metadata.Version)
		} else {
			if err := c.HelmClient.Package(r.Release.Chart.Metadata.Name, r.Release.Chart.Metadata.Version); err != nil {
				zap.L().Sugar().Infof("Can't package %s-%s chart in local filesystem: %v", r.Release.Chart.Metadata.Name, r.Release.Chart.Metadata.Version, err)
				continue
			}
			r.IsPackaged = true
		}

		if c.ChartmuseumClient.IsActive() {
			packageFile, err := c.HelmClient.GetReleasePackageFile(r)
			if err != nil {
				zap.L().Sugar().Infof("Can't get package file for %s-%s chart in local filesystem: %v", r.Release.Chart.Metadata.Name, r.Release.Chart.Metadata.Version, err)
				continue
			}

			err = c.ChartmuseumClient.Upload(r.Release.Chart.Metadata.Name, r.Release.Chart.Metadata.Version, packageFile)
			if err != nil {
				zap.L().Sugar().Infof("Can't upload %s-%s chart: %v", r.Release.Chart.Metadata.Name, r.Release.Chart.Metadata.Version, err)
				continue
			}
		}
	}

	return nil
}
