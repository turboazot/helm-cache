package services

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"mime/multipart"
	"net/http"
	"os"
	"time"

	"github.com/hashicorp/go-retryablehttp"
	"github.com/turboazot/helm-cache/pkg/entities"

	"go.uber.org/zap"
)

type ChartmuseumClient struct {
	ChartmuseumUrl      string
	ChartmuseumUsername string
	ChartmuseumPassword string
	HttpClient          *retryablehttp.Client
	ChartVersionCache   map[string]bool
}

func NewChartmuseumClient(chartmuseumUrl string, chartmuseumUsername string, chartmuseumPassword string) (*ChartmuseumClient, error) {
	retryClient := retryablehttp.NewClient()
	retryClient.RetryMax = 5
	retryClient.HTTPClient.Timeout = 5 * time.Second

	var c *ChartmuseumClient = &ChartmuseumClient{
		ChartmuseumUrl:      chartmuseumUrl,
		ChartmuseumUsername: chartmuseumUsername,
		ChartmuseumPassword: chartmuseumPassword,
		HttpClient:          retryClient,
		ChartVersionCache:   make(map[string]bool),
	}

	if !c.IsActive() {
		return c, nil
	}

	chartListBytes, err := c.GetAllCharts()
	if err != nil {
		return nil, err
	}

	var chartsMap map[string][]entities.RestChart

	err = json.Unmarshal(chartListBytes, &chartsMap)
	if err != nil {
		return nil, err
	}

	for chartName, chartsArray := range chartsMap {
		for _, chart := range chartsArray {
			chartInstanceVersion := chart.Version
			c.ChartVersionCache[fmt.Sprintf("%s-%s", chartName, chartInstanceVersion)] = true
		}
	}

	return c, nil
}

func (c *ChartmuseumClient) IsActive() bool {
	return c.ChartmuseumUrl != ""
}

func (c *ChartmuseumClient) hasBasicAuth() bool {
	return c.ChartmuseumUsername != "" && c.ChartmuseumPassword != ""
}

func (c *ChartmuseumClient) GetAllCharts() ([]byte, error) {
	req, err := retryablehttp.NewRequest("GET", fmt.Sprintf("%s/api/charts", c.ChartmuseumUrl), nil)
	if err != nil {
		return nil, err
	}

	if c.hasBasicAuth() {
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
	req, err := retryablehttp.NewRequest("POST", fmt.Sprintf("%s/api/charts", c.ChartmuseumUrl), body)
	if err != nil {
		return err
	}

	if c.hasBasicAuth() {
		req.SetBasicAuth(c.ChartmuseumUsername, c.ChartmuseumPassword)
	}
	req.Header.Add("Content-Type", writer.FormDataContentType())

	resp, err := c.HttpClient.Do(req)
	if err != nil {
		return err
	}

	responseBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	if resp.StatusCode != http.StatusCreated {
		return errors.New(fmt.Sprintf("Receiving list of charts failed. Status code - %d, Body - %s", resp.StatusCode, string(responseBody)))
	}

	c.ChartVersionCache[fmt.Sprintf("%s-%s", chartName, chartVersion)] = true

	zap.L().Sugar().Infof("Successfully uploaded chart: %s-%s", chartName, chartVersion)

	return resp.Body.Close()
}

func (c *ChartmuseumClient) IsExists(chartName string, chartVersion string) bool {
	return c.ChartVersionCache[fmt.Sprintf("%s-%s", chartName, chartVersion)]
}
