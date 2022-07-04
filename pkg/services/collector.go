package services

import (
	"archive/tar"
	"bytes"
	"compress/gzip"
	"context"
	"encoding/base64"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"mime/multipart"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/Jeffail/gabs"
	"gopkg.in/yaml.v2"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/utils/strings/slices"
)

func isDirectory(path string) (bool, error) {
	fileInfo, err := os.Stat(path)
	if err != nil {
		return false, err
	}

	return fileInfo.IsDir(), err
}

func writeStringToFile(path string, content string) error {
	directory := filepath.Dir(path)

	if directory != "." {
		os.MkdirAll(directory, 0755)
	}

	f, err := os.Create(path)
	if err != nil {
		return err
	}
	defer f.Close()

	_, err = f.WriteString(content)
	if err != nil {
		return err
	}

	return nil
}

func writeMapContainerToFile(path string, container *gabs.Container) error {
	m := container.Data()
	d, err := yaml.Marshal(&m)
	if err != nil {
		return err
	}
	content := string(d)

	err = writeStringToFile(path, content)
	if err != nil {
		return err
	}

	return nil
}

func createChartFiles(chartPath string, containers []*gabs.Container) error {
	for _, container := range containers {
		containerMap, err := container.ChildrenMap()
		if err != nil {
			return err
		}

		fileName := containerMap["name"].Data().(string)
		fileData := containerMap["data"].Data().(string)
		fileDataBytes, err := base64.StdEncoding.DecodeString(fileData)
		if err != nil {
			return err
		}
		fileDataString := string(fileDataBytes)
		err = writeStringToFile(fmt.Sprintf("%s/%s", chartPath, fileName), fileDataString)
		if err != nil {
			return err
		}
	}
	return nil
}

func createArchive(files []string, replaceOldPath string, replaceNewPath string, buf io.Writer) error {
	// Create new Writers for gzip and tar
	// These writers are chained. Writing to the tar writer will
	// write to the gzip writer which in turn will write to
	// the "buf" writer
	gw := gzip.NewWriter(buf)
	defer gw.Close()
	tw := tar.NewWriter(gw)
	defer tw.Close()

	// Iterate over files and add them to the tar archive
	for _, file := range files {
		err := addToArchive(tw, file, replaceOldPath, replaceNewPath)
		if err != nil {
			return err
		}
	}

	return nil
}

func addToArchive(tw *tar.Writer, filename string, replaceOldPath string, replaceNewPath string) error {
	// Open the file which will be written into the archive
	file, err := os.Open(filename)
	if err != nil {
		return err
	}
	defer file.Close()

	// Get FileInfo about our file providing file size, mode, etc.
	info, err := file.Stat()
	if err != nil {
		return err
	}

	// Create a tar Header from the FileInfo data
	header, err := tar.FileInfoHeader(info, info.Name())
	if err != nil {
		return err
	}

	// Use full path as name (FileInfoHeader only takes the basename)
	// If we don't do this the directory strucuture would
	// not be preserved
	// https://golang.org/src/archive/tar/common.go?#L626
	filename = strings.Replace(filename, replaceOldPath, replaceNewPath, 1)
	header.Name = filename

	// Write file header to the tar archive
	err = tw.WriteHeader(header)
	if err != nil {
		return err
	}

	// Copy file content to tar archive
	_, err = io.Copy(tw, file)
	if err != nil {
		return err
	}

	return nil
}

func execCommand(command string) (string, error) {
	output, err := exec.Command("sh", "-c", command).CombinedOutput()
	outputString := string(output)
	fmt.Printf("CMD: %s;\nOutput:\n%s\n", command, outputString)
	return outputString, err
}

func getChartNameAndRevisionFromSecretName(secretName string) (string, int, error) {
	secretNameSplitted := strings.Split(secretName, ".")
	secretChartName := secretNameSplitted[4]
	secretRevisionInt, err := strconv.Atoi(strings.Replace(secretNameSplitted[5], "v", "", -1))
	return secretChartName, secretRevisionInt, err
}

type Collector struct {
	ChartsRootDir             string
	ChartmuseumUrl            string
	ChartmuseumUsername       string
	ChartmuseumPassword       string
	KubernetesClientset       *kubernetes.Clientset
	HttpClient                *http.Client
	ChartVersionsRegistry     map[string][]string
	ChartLastRevisionRegistry map[string]int
}

func NewCollector(chartmuseumUrl string, chartmuseumUsername string, chartmuseumPassword string) *Collector {
	// In-cluster config
	config, err := rest.InClusterConfig()
	if err != nil {
		panic(err.Error())
	}

	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		panic(err.Error())
	}

	return &Collector{
		ChartsRootDir:             "./parsed-charts",
		ChartmuseumUrl:            chartmuseumUrl,
		ChartmuseumUsername:       chartmuseumUsername,
		ChartmuseumPassword:       chartmuseumPassword,
		ChartVersionsRegistry:     make(map[string][]string),
		ChartLastRevisionRegistry: make(map[string]int),
		HttpClient:                &http.Client{},
		KubernetesClientset:       clientset,
	}
}

func (c *Collector) RefreshChartVersionRegistry(secrets *v1.SecretList) {
	// Get all helm charts
	req, err := http.NewRequest("GET", fmt.Sprintf("%s/api/charts", c.ChartmuseumUrl), nil)
	req.SetBasicAuth(c.ChartmuseumUsername, c.ChartmuseumPassword)
	resp, err := c.HttpClient.Do(req)
	if err != nil {
		panic(err)
	}

	defer resp.Body.Close()
	responseBody, err := io.ReadAll(resp.Body)
	if resp.StatusCode != 200 {
		log.Fatalf("Receiving list of charts failed. Status code - %d, Body - %s", resp.StatusCode, string(responseBody))
		os.Exit(1)
	}

	chartListContainer, err := gabs.ParseJSON(responseBody)
	if err != nil {
		panic(err)
	}

	chartsMap, err := chartListContainer.ChildrenMap()
	if err != nil {
		panic(err)
	}

	for name, _ := range chartsMap {
		chartsInstances, err := chartListContainer.Path(name).Children()
		if err != nil {
			log.Fatal(err)
			continue
		}

		for _, chartInstance := range chartsInstances {
			chartInstanceMap, err := chartInstance.ChildrenMap()
			if err != nil {
				log.Fatal(err)
				continue
			}

			chartInstanceVersion := chartInstanceMap["version"].Data().(string)
			if !slices.Contains(c.ChartVersionsRegistry[name], chartInstanceVersion) {
				c.ChartVersionsRegistry[name] = append(c.ChartVersionsRegistry[name], chartInstanceVersion)
			}
		}
	}
}

func (c *Collector) RefreshChartLastRevisionRegistry(secrets *v1.SecretList) {
	for _, secret := range secrets.Items {
		if !strings.HasPrefix(secret.Name, "sh.helm.release.v1") {
			continue
		}
		secretChartName, secretRevisionInt, err := getChartNameAndRevisionFromSecretName(secret.Name)
		if err != nil {
			log.Fatal(err)
			continue
		}

		c.ChartLastRevisionRegistry[secretChartName] = secretRevisionInt
	}
}

func (c *Collector) DecodeReleaseString(release string) (*gabs.Container, error) {
	encodedRelease, err := base64.StdEncoding.DecodeString(release)
	if err != nil {
		return nil, err
	}
	decodedReleaseReader, err := gzip.NewReader(bytes.NewReader(encodedRelease))
	if err != nil {
		return nil, err
	}
	decodedReleaseBytes, err := ioutil.ReadAll(decodedReleaseReader)
	if err != nil {
		return nil, err
	}

	return gabs.ParseJSON(decodedReleaseBytes)
}

func (c *Collector) PackageHelmChart(chartName string, chartVersion string, chartPath string) error {
	var packageFiles []string

	err := filepath.Walk(chartPath,
		func(path string, info os.FileInfo, err error) error {
			if err != nil {
				return err
			}
			pathIsDirectory, err := isDirectory(path)
			if err != nil {
				return err
			}
			if !pathIsDirectory {
				packageFiles = append(packageFiles, path)
			}

			return nil
		})
	if err != nil {
		return err
	}

	out, err := os.Create(fmt.Sprintf("%s-%s.tgz", chartName, chartVersion))
	if err != nil {
		return err
	}
	defer out.Close()

	// Create the archive and write the output to the "out" Writer
	err = createArchive(packageFiles, fmt.Sprintf("%s/%s/", filepath.Dir(chartPath), filepath.Base(chartPath)), fmt.Sprintf("%s/", chartName), out)
	return err
}

func (c *Collector) AddHelmRepositories(releaseContainer *gabs.Container) error {
	if !releaseContainer.ExistsP("chart.metadata.dependencies") {
		return nil
	}
	dependencies, err := releaseContainer.Path("chart.metadata.dependencies").Children()
	if err != nil {
		return err
	}

	var repositoryNames []string

	for _, dependency := range dependencies {
		dependencyMap, err := dependency.ChildrenMap()
		if err != nil {
			return err
		}

		repositoryName := dependencyMap["name"].Data().(string)
		repositoryUrl := dependencyMap["repository"].Data().(string)

		_, err = execCommand(fmt.Sprintf("helm repo add %s %s", repositoryName, repositoryUrl))
		if err != nil {
			return err
		}

		repositoryNames = append(repositoryNames, repositoryName)
	}

	_, err = execCommand(fmt.Sprintf("helm repo update %s", strings.Join(repositoryNames, " ")))
	return err
}

func (c *Collector) UploadHelmPackage(chartName string, chartVersion string) error {
	file, err := os.Open(fmt.Sprintf("%s-%s.tgz", chartName, chartVersion))
	if err != nil {
		return err
	}
	fileContents, err := ioutil.ReadAll(file)
	if err != nil {
		return err
	}
	fi, err := file.Stat()
	if err != nil {
		return err
	}
	file.Close()

	body := new(bytes.Buffer)
	writer := multipart.NewWriter(body)
	part, err := writer.CreateFormFile("chart", fi.Name())
	if err != nil {
		return err
	}
	part.Write(fileContents)
	writer.Close()
	req, err := http.NewRequest("POST", fmt.Sprintf("%s/api/charts", c.ChartmuseumUrl), body)
	if err != nil {
		return err
	}
	req.SetBasicAuth(c.ChartmuseumUsername, c.ChartmuseumPassword)
	req.Header.Add("Content-Type", writer.FormDataContentType())
	resp, err := c.HttpClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	responseBody, err := io.ReadAll(resp.Body)
	if resp.StatusCode != 201 {
		return errors.New(fmt.Sprintf("Receiving list of charts failed. Status code - %d, Body - %s", resp.StatusCode, string(responseBody)))
	}
	c.ChartVersionsRegistry[chartName] = append(c.ChartVersionsRegistry[chartName], chartVersion)
	fmt.Printf("Chart %s-%s has been saved successfully\n", chartName, chartVersion)
	return nil
}

func (c *Collector) CheckAllSecrets() {
	secrets, err := c.KubernetesClientset.CoreV1().Secrets("").List(context.TODO(), metav1.ListOptions{})
	if err != nil {
		panic(err.Error())
	}

	c.RefreshChartVersionRegistry(secrets)

	// Remove old versions
	c.RefreshChartLastRevisionRegistry(secrets)

	for _, secret := range secrets.Items {
		if !strings.HasPrefix(secret.Name, "sh.helm.release.v1") {
			continue
		}
		secretChartName, secretRevisionInt, err := getChartNameAndRevisionFromSecretName(secret.Name)
		if err != nil {
			log.Fatal(err)
			continue
		}
		if secretRevisionInt < c.ChartLastRevisionRegistry[secretChartName] {
			continue
		}

		fmt.Printf("Checking secret %s...\n", secret.Name)

		decodedReleaseContainer, err := c.DecodeReleaseString(string(secret.Data["release"]))
		if err != nil {
			log.Fatal(err)
			continue
		}

		chartName := decodedReleaseContainer.Path("chart.metadata.name").Data().(string)
		chartVersion := decodedReleaseContainer.Path("chart.metadata.version").Data().(string)
		chartPath := fmt.Sprintf("%s/%s-%s", c.ChartsRootDir, chartName, chartVersion)

		// Check if chart is already in registry
		if slices.Contains(c.ChartVersionsRegistry[chartName], chartVersion) {
			fmt.Printf("Chart %s-%s already exists. Skipping\n", chartName, chartVersion)
			os.RemoveAll(chartPath)
			os.RemoveAll(fmt.Sprintf("%s-%s.tgz", chartName, chartVersion))
			continue
		}

		// Create values.yaml
		valuesContainer := decodedReleaseContainer.Path("chart.values")
		err = writeMapContainerToFile(fmt.Sprintf("%s/values.yaml", chartPath), valuesContainer)
		if err != nil {
			log.Fatal(err)
			continue
		}

		// Create Chart.yaml
		chartYamlContainer := decodedReleaseContainer.Path("chart.metadata")
		err = writeMapContainerToFile(fmt.Sprintf("%s/Chart.yaml", chartPath), chartYamlContainer)
		if err != nil {
			log.Fatal(err)
			continue
		}

		// Create templates
		templates, err := decodedReleaseContainer.Path("chart.templates").Children()
		if err != nil {
			log.Fatal(err)
			continue
		}
		err = createChartFiles(chartPath, templates)
		if err != nil {
			log.Fatal(err)
			continue
		}

		// Create files
		files, err := decodedReleaseContainer.Path("chart.files").Children()
		if err != nil {
			log.Fatal(err)
			continue
		}
		err = createChartFiles(chartPath, files)
		if err != nil {
			log.Fatal(err)
			continue
		}

		// Add repositories
		err = c.AddHelmRepositories(decodedReleaseContainer)
		if err != nil {
			log.Fatal(err)
			continue
		}

		// Fetch dependencies
		_, err = execCommand(fmt.Sprintf("helm dep update %s", chartPath))
		if err != nil {
			log.Fatal(err)
			continue
		}

		// Create package
		err = c.PackageHelmChart(chartName, chartVersion, chartPath)
		if err != nil {
			log.Fatal(err)
			continue
		}

		// Upload helm package
		err = c.UploadHelmPackage(chartName, chartVersion)
		if err != nil {
			log.Fatal(err)
			continue
		}
		os.RemoveAll(chartPath)
		os.RemoveAll(fmt.Sprintf("%s-%s.tgz", chartName, chartVersion))
	}
}
