package mockserver

import (
	"log"
	"os"
	"fmt"
	"net/http"
	"encoding/json"
	"bytes"
)

type Client struct {
	mockBaseUrl string
	proxyBaseUrl string
	httpClient  *http.Client
	log         *log.Logger
}

func NewClient(mockBaseUrl, proxyBaseUrl string) *Client {
	return &Client{
		mockBaseUrl: mockBaseUrl,
		proxyBaseUrl: proxyBaseUrl,
		httpClient: &http.Client{},
		log: log.New(os.Stdout, fmt.Sprintf("mockserver(%v/%v): ", mockBaseUrl, proxyBaseUrl), 0),
	}
}

func (c *Client) MockDo(path string, requestBody interface{}) error {
	return c.do(c.mockBaseUrl, path, requestBody)
}

func (c *Client) ProxyDo(path string, requestBody interface{}) error {
	return c.do(c.proxyBaseUrl, path, requestBody)
}

func (c *Client) MockReset() error {
	return c.MockDo("/reset", nil)
}

func (c *Client) ProxyReset() error {
	return c.ProxyDo("/reset", nil)
}

func (c *Client) do(baseUrl, path string, requestBody interface{}) error {
	url := fmt.Sprintf("%v%v", baseUrl, path)
	c.log.Printf("sending request %v(%T)", path, requestBody)

	var bodyReader *bytes.Buffer

	if requestBody != nil {
		serializedBody, err := json.Marshal(requestBody)
		if err != nil {
			return err
		}
		bodyReader = bytes.NewBuffer(serializedBody)
	} else {
		bodyReader = bytes.NewBuffer([]byte{})
	}

	req, err := http.NewRequest("PUT", url, bodyReader)
	if err != nil {
		return err
	}
	req.Header.Add("Content-Type", "application/json; charset=utf-8")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return err
	}
	if resp.StatusCode < 200 || resp.StatusCode > 299 {
		return fmt.Errorf("request failed with status %v", resp.StatusCode)
	}
	return nil
}