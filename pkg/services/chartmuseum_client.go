package services

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"mime/multipart"
	"net/http"
	"os"

	"github.com/Jeffail/gabs"
	"go.uber.org/zap"
)

type ChartmuseumClient struct {
	ChartmuseumUrl      string
	ChartmuseumUsername string
	ChartmuseumPassword string
	HttpClient          *http.Client
	ChartVersionCache   map[string]map[string]bool
}

func NewChartmuseumClient(chartmuseumUrl string, chartmuseumUsername string, chartmuseumPassword string) (*ChartmuseumClient, error) {
	var c *ChartmuseumClient = &ChartmuseumClient{
		ChartmuseumUrl:      chartmuseumUrl,
		ChartmuseumUsername: chartmuseumUsername,
		ChartmuseumPassword: chartmuseumPassword,
		HttpClient:          &http.Client{},
		ChartVersionCache:   make(map[string]map[string]bool),
	}

	if !c.IsActive() {
		return c, nil
	}

	chartListBytes, err := c.GetAllCharts()
	chartListContainer, err := gabs.ParseJSON(chartListBytes)
	if err != nil {
		return nil, err
	}

	chartsMapContainer, err := chartListContainer.ChildrenMap()
	if err != nil {
		return nil, err
	}

	for chartName, _ := range chartsMapContainer {
		chartsInstancesContainer, err := chartListContainer.Path(chartName).Children()
		if err != nil {
			return nil, err
		}

		for _, chartInstanceContainer := range chartsInstancesContainer {
			chartInstanceContainerMap, err := chartInstanceContainer.ChildrenMap()
			if err != nil {
				return nil, err
			}

			chartInstanceVersion := chartInstanceContainerMap["version"].Data().(string)
			if _, chartInstanceExists := c.ChartVersionCache[chartName]; !chartInstanceExists {
				c.ChartVersionCache[chartName] = make(map[string]bool)
			}
			c.ChartVersionCache[chartName][chartInstanceVersion] = true
		}
	}

	return c, nil
}

func (c *ChartmuseumClient) IsActive() bool {
	return c.ChartmuseumUrl != ""
}

func (c *ChartmuseumClient) HasBasicAuth() bool {
	return c.ChartmuseumUsername != "" && c.ChartmuseumPassword != ""
}

func (c *ChartmuseumClient) GetAllCharts() ([]byte, error) {
	req, err := http.NewRequest("GET", fmt.Sprintf("%s/api/charts", c.ChartmuseumUrl), nil)
	if err != nil {
		return nil, err
	}

	if c.HasBasicAuth() {
		req.SetBasicAuth(c.ChartmuseumUsername, c.ChartmuseumPassword)
	}

	resp, err := c.HttpClient.Do(req)
	if err != nil {
		return nil, err
	}

	defer resp.Body.Close()

	respBody, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode != 200 {
		return nil, errors.New(fmt.Sprintf("Receiving list of charts failed. Status code - %d, Body - %s", resp.StatusCode, string(respBody)))
	}

	return respBody, nil
}

func (c *ChartmuseumClient) Upload(chartName string, chartVersion string, f *os.File) error {
	// file, err := os.Open("/Users/turboazot/projects/ln/helm-cache/home/data/packaged/nginx-12.0.5.tgz")
	// if err != nil {
	// 	return err
	// }
	// fileContents, err := ioutil.ReadAll(file)
	// if err != nil {
	// 	return err
	// }
	// fi, err := file.Stat()
	// if err != nil {
	// 	return err
	// }
	// file.Close()

	// body := new(bytes.Buffer)
	// writer := multipart.NewWriter(body)
	// part, err := writer.CreateFormFile("chart", fi.Name())
	// if err != nil {
	// 	return err
	// }
	// part.Write(fileContents)
	// writer.Close()
	// req, err := http.NewRequest("POST", fmt.Sprintf("%s/api/charts", c.ChartmuseumUrl), body)
	// if err != nil {
	// 	return err
	// }
	// req.SetBasicAuth(c.ChartmuseumUsername, c.ChartmuseumPassword)
	// req.Header.Add("Content-Type", writer.FormDataContentType())
	// resp, err := c.HttpClient.Do(req)
	// if err != nil {
	// 	return err
	// }
	// defer resp.Body.Close()
	// fmt.Println(resp.StatusCode)
	// responseBody, err := io.ReadAll(resp.Body)
	// if resp.StatusCode != 201 {
	// 	return errors.New(fmt.Sprintf("Receiving list of charts failed. Status code - %d, Body - %s", resp.StatusCode, string(responseBody)))
	// }

	// ---------------

	// // Request creation
	// req, err := http.NewRequest("POST", fmt.Sprintf("%s/api/charts", c.ChartmuseumUrl), nil)
	// if err != nil {
	// 	return err
	// }

	// // Rest
	// var body bytes.Buffer
	// // chartPackagePath := fmt.Sprintf("%s/%s-%s.tgz", "/Users/turboazot/.helm-cache/data/packaged", "kube-prometheus-stack", "36.2.0")
	// chartPackagePath := "/Users/turboazot/projects/ln/helm-cache/home/data/packaged/nginx-12.0.5.tgz"
	// w := multipart.NewWriter(&body)
	// defer w.Close()
	// fw, err := w.CreateFormFile("chart", chartPackagePath)
	// if err != nil {
	// 	return err
	// }
	// w.FormDataContentType()
	// fd, err := os.Open(chartPackagePath)
	// if err != nil {
	// 	return err
	// }
	// defer fd.Close()
	// l, err := io.Copy(fw, fd)
	// if err != nil {
	// 	return err
	// }

	// fmt.Println("Content length", l)
	// req.Header.Set("Content-Type", w.FormDataContentType())
	// req.Header.Set("Content-Length", "39715")
	// if c.HasBasicAuth() {
	// 	req.SetBasicAuth(c.ChartmuseumUsername, c.ChartmuseumPassword)
	// }
	// dump, _ := httputil.DumpRequest(req, true)

	// fmt.Println(string(dump))
	// req.Body = ioutil.NopCloser(&body)
	// resp, err := c.HttpClient.Do(req)
	// fmt.Println(resp.StatusCode)
	// return nil

	// --------------------------

	// f, err := os.Open("/Users/turboazot/projects/ln/helm-cache/home/data/packaged/nginx-12.0.5.tgz")
	// if err != nil {
	// 	return err
	// }

	fileContents, err := ioutil.ReadAll(f)
	if err != nil {
		return err
	}
	fi, err := f.Stat()
	if err != nil {
		return err
	}
	f.Close()
	if err != nil {
		return err
	}

	body := new(bytes.Buffer)
	writer := multipart.NewWriter(body)
	part, err := writer.CreateFormFile("chart", fi.Name())
	if err != nil {
		return err
	}
	_, err = part.Write(fileContents)
	if err != nil {
		return err
	}
	err = writer.Close()
	if err != nil {
		return err
	}
	req, err := http.NewRequest("POST", fmt.Sprintf("%s/api/charts", c.ChartmuseumUrl), body)
	if err != nil {
		return err
	}

	if c.HasBasicAuth() {
		req.SetBasicAuth(c.ChartmuseumUsername, c.ChartmuseumPassword)
	}
	req.Header.Add("Content-Type", writer.FormDataContentType())
	fmt.Println("gg")
	resp, err := c.HttpClient.Do(req)
	if err != nil {
		return err
	}
	fmt.Println("wp")
	responseBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}
	fmt.Println(resp)
	if resp.StatusCode != 201 {
		return errors.New(fmt.Sprintf("Receiving list of charts failed. Status code - %d, Body - %s", resp.StatusCode, string(responseBody)))
		// return errors.New(fmt.Sprintf("Receiving list of charts failed. Status code - %d, Body - %s", resp.StatusCode, "Huy"))
	}

	if _, chartExists := c.ChartVersionCache[chartName]; !chartExists {
		c.ChartVersionCache[chartName] = make(map[string]bool)
	}

	c.ChartVersionCache[chartName][chartVersion] = true

	zap.L().Sugar().Infof("Successfully uploaded chart: %s-%s", chartName, chartVersion)

	return resp.Body.Close()
}

func (c *ChartmuseumClient) IsExists(chartName string, chartVersion string) bool {
	if _, chartExists := c.ChartVersionCache[chartName]; !chartExists {
		return false
	}
	return c.ChartVersionCache[chartName][chartVersion]
}
