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
	baseUrl string
	httpClient *http.Client
	log *log.Logger
}

func NewClient(baseUrl string) *Client {
	return &Client{
		baseUrl: baseUrl,
		httpClient: &http.Client{},
		log: log.New(os.Stdout, fmt.Sprintf("mockserver(%v): ", baseUrl), 0),
	}
}

func (c *Client) Do(path string, requestBody interface{}) error {
	c.log.Printf("sending request %v(%T)", path, requestBody)

	serializedBody, err := json.Marshal(requestBody)
	if err != nil {
		return err
	}

	req, err := http.NewRequest("PUT", fmt.Sprintf("%v%v", c.baseUrl, path), bytes.NewBuffer(serializedBody))
	if err != nil {
		return err
	}
	req.Header.Add("Content-Type", "application/json; charset=utf-8")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return err
	}
	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
		return fmt.Errorf("request failed with status %v", resp.StatusCode)
	}
	return nil
}
