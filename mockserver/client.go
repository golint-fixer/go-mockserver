package mockserver

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
)

type Client struct {
	mockBaseURL  string
	proxyBaseURL string
	httpClient   *http.Client
	log          *log.Logger
}

func NewClient(mockBaseUrl, proxyBaseUrl string) *Client {
	return &Client{
		mockBaseURL:  mockBaseUrl,
		proxyBaseURL: proxyBaseUrl,
		httpClient:   &http.Client{},
		log:          log.New(os.Stdout, fmt.Sprintf("mockserver(%v/%v): ", mockBaseUrl, proxyBaseUrl), 0),
	}
}

func (c *Client) GetMockURL(path string) string {
	return fmt.Sprintf("%v%v", c.mockBaseURL, path)
}

func (c *Client) GetProxyURL(path string) string {
	return fmt.Sprintf("%v%v", c.proxyBaseURL, path)
}

func (c *Client) MockAnyResponse(mockAnyResponse *MockAnyResponse) error {
	_, err := c.mockDo("/expectation", mockAnyResponse)
	return err
}

func (c *Client) MustMockAnyResponse(mockAnyResponse *MockAnyResponse) {
	if err := c.MockAnyResponse(mockAnyResponse); err != nil {
		panic(err)
	}
}

func (c *Client) ResetMocks() error {
	_, err := c.mockDo("/reset", nil)
	return err
}

func (c *Client) MustResetMocks() {
	if err := c.ResetMocks(); err != nil {
		panic(err)
	}
}

func (c *Client) VerifyProxy(verify *Verify) error {
	_, err := c.proxyDo("/verify", verify)
	return err
}

func (c *Client) MustVerifyProxy(verify *Verify) {
	if err := c.VerifyProxy(verify); err != nil {
		panic(err)
	}
}

func (c *Client) RetrieveProxy(retrieve *Retrieve) ([]*RetrievedRequest, error) {
	respBody, err := c.proxyDo("/retrieve", retrieve.HttpRequest)
	if err != nil {
		return nil, err
	}
	requests := make([]*RetrievedRequest, 0)
	if err := json.Unmarshal(respBody, &requests); err != nil {
		return nil, err
	}
	return requests, nil
}

func (c *Client) MustRetrieveProxy(retrieve *Retrieve) []*RetrievedRequest {
	if requests, err := c.RetrieveProxy(retrieve); err == nil {
		return requests
	} else {
		panic(err)
	}
}

func (c *Client) ResetProxy() error {
	_, err := c.proxyDo("/reset", nil)
	return err
}

func (c *Client) MustResetProxy() {
	if err := c.ResetProxy(); err != nil {
		panic(err)
	}
}

func (c *Client) mockDo(path string, requestBody interface{}) ([]byte, error) {
	return c.do(c.mockBaseURL, path, requestBody)
}

func (c *Client) proxyDo(path string, requestBody interface{}) ([]byte, error) {
	return c.do(c.proxyBaseURL, path, requestBody)
}

func (c *Client) do(baseUrl, path string, requestBody interface{}) ([]byte, error) {
	url := fmt.Sprintf("%v%v", baseUrl, path)
	c.log.Printf("sending request %v(%T)", url, requestBody)
	var bodyReader *bytes.Buffer
	if requestBody != nil {
		serializedBody, err := json.Marshal(requestBody)
		if err != nil {
			return nil, err
		}
		bodyReader = bytes.NewBuffer(serializedBody)
		c.log.Printf("request body %s\n", serializedBody)
	} else {
		bodyReader = bytes.NewBuffer([]byte{})
	}
	req, err := http.NewRequest("PUT", url, bodyReader)
	if err != nil {
		return nil, err
	}
	req.Header.Add("Content-Type", "application/json; charset=utf-8")
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode < 200 || resp.StatusCode > 299 {
		return nil, fmt.Errorf("request failed with status %v", resp.StatusCode)
	}
	respBody, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	c.log.Printf("mockserver response: %s", respBody)
	return respBody, nil
}
