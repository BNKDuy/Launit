package api

import (
	"bytes"
	"encoding/json"
	"errors"
	"net/http"
	"os"
)

type Client struct {
	baseUrl    string
	httpClient *http.Client
}

func NewClient(baseUrl string) *Client {
	return &Client{
		baseUrl:    baseUrl,
		httpClient: &http.Client{},
	}
}

type presignedUrlResponse struct {
	Url string `json:"1"`
}

func newPresignedUrlResponse(url string) *presignedUrlResponse {
	return &presignedUrlResponse{
		Url: url,
	}
}

func (c *Client) GetPresignedURL(functionName string) (string, error) {
	uploadRequest := newUploadRequest(functionName)
	body, err := json.Marshal(uploadRequest)
	if err != nil {
		return "", errors.New("failed to encode upload metadata: " + err.Error())
	}

	req, err := http.NewRequest("POST", c.baseUrl+ComputeUploadEndpoint, bytes.NewBuffer(body))
	if err != nil {
		return "", errors.New("failed to construct upload request: " + err.Error())
	}
	req.Header.Set("Content-Type", "application/json")

	res, err := c.httpClient.Do(req)
	if err != nil {
		return "", errors.New("failed to request an upload link from the orchestrator: " + err.Error())
	}
	if res != nil && res.Body != nil {
		defer res.Body.Close()
	}

	if res.StatusCode != http.StatusOK {
		return "", errors.New("orchestrator returned status code: " + res.Status)
	}

	var payload presignedUrlResponse
	if err := json.NewDecoder(res.Body).Decode(&payload); err != nil {
		return "", errors.New("failed to decode orchestrator response: " + err.Error())
	}

	return payload.Url, nil
}

func (c *Client) UploadZip(presignedUrl, zipPath string) error {
	stat, err := os.Stat(zipPath)
	if err != nil {
		if os.IsNotExist(err) {
			return errors.New("Cannot locate zip file")
		}
		return errors.New("Fail to retrieve information about zip file")
	}
	size := stat.Size()

	file, err := os.Open(zipPath)
	if err != nil {
		return errors.New("Fail to open zip file")
	}
	defer file.Close()

	req, err := http.NewRequest("PUT", presignedUrl, file)
	if err != nil {
		return errors.New("Failed to construct S3 storage upload request: " + err.Error())
	}

	req.ContentLength = size
	req.Header.Set("Content-Type", "application/zip")

	res, err := c.httpClient.Do(req)
	if err != nil {
		return errors.New("Failed to transmit data to storage node: " + err.Error())
	}
	defer res.Body.Close()

	if res.StatusCode != http.StatusOK {
		return errors.New("Storage server rejected the upload package with status: " + res.Status)
	}

	return nil
}

type CreateResponse struct {
	Url string `json:"1"`
}

func newCreateResponse(url string) *CreateResponse {
	return &CreateResponse{
		Url: url,
	}
}

func (c *Client) SendCreateRequest(functionName, runtime, memory string, timeout int32) (string, error) {
	body := newCreateRequest(functionName, memory, runtime, timeout)
	reqBody, err := json.Marshal(body)
	if err != nil {
		return "", errors.New("Failed to encode create request: " + err.Error())
	}

	req, err := http.NewRequest("POST", c.baseUrl+ComputeCreateEndpoint, bytes.NewBuffer(reqBody))
	if err != nil {
		return "", errors.New("Faield to construct create request: " + err.Error())
	}

	res, err := c.httpClient.Do(req)
	if err != nil {
		return "", errors.New("failed to request the orchestrator to create a new function: " + err.Error())
	}
	if res != nil && res.Body != nil {
		defer res.Body.Close()
	}

	if res.StatusCode != http.StatusCreated {
		return "", errors.New("orchestrator returned status code: " + res.Status)
	}

	var payload CreateResponse
	if err := json.NewDecoder(res.Body).Decode(&payload); err != nil {
		return "", errors.New("failed to decode orchestrator response: " + err.Error())
	}
	if payload.Url == "" {
		return "", errors.New("Failed to get URL for the newly created function.")
	}

	return payload.Url, nil
}
