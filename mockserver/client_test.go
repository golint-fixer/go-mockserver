package mockserver_test

import (
	"bytes"
	"crypto/tls"
	"fmt"
	"github.com/ibrt/go-compose/compose"
	"github.com/ibrt/go-mockserver/mockserver"
	"github.com/stretchr/testify/assert"
	"net/http"
	"net/url"
	"testing"
)

var composeYML = `
mockserver:
  container_name: mockserver
  image: jamesdbloom/mockserver
  ports:
    - "1080"
    - "1090"
`

func TestMockAnyResponse(t *testing.T) {
	compose := compose.MustStart(composeYML, true, true)
	defer compose.Kill()
	client := newClient(compose.Containers["mockserver"])

	err := client.MockAnyResponse(
		mockserver.NewMockAnyResponse().
			When(mockserver.NewRequest("GET", "/test")).
			Respond(mockserver.NewResponse(http.StatusOK)))
	assert.Nil(t, err)

	resp, err := http.Get(client.GetMockURL("/test"))
	assert.Nil(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestMockReset(t *testing.T) {
	compose := compose.MustStart(composeYML, true, true)
	defer compose.Kill()
	client := newClient(compose.Containers["mockserver"])

	assert.Nil(t, client.ResetMocks())
}

func TestProxy(t *testing.T) {
	compose := compose.MustStart(composeYML, true, true)
	defer compose.Kill()
	client := newClient(compose.Containers["mockserver"])
	httpClient := newHttpClient(client.GetProxyURL(""))

	_, err := httpClient.Get("https://www.google.com/")
	assert.Nil(t, err)

	_, err = httpClient.Post("http://www.google.com/", "application/json", bytes.NewBuffer([]byte("{ \"hello\": true }")))
	assert.Nil(t, err)

	err = client.VerifyProxy(
		mockserver.NewVerify().
			MatchRequest(mockserver.NewRequest("GET", "/")).
			WithTimes(1, true))
	assert.Nil(t, err)

	_, err = client.RetrieveProxy(
		mockserver.NewRetrieve().
			MatchRequest(mockserver.NewRequest("GET", "/")))
	assert.Nil(t, err)

	_, err = client.RetrieveProxy(
		mockserver.NewRetrieve().
			MatchRequest(mockserver.NewRequest("POST", "/")))
	assert.Nil(t, err)
}

func TestProxyReset(t *testing.T) {
	compose := compose.MustStart(composeYML, true, true)
	defer compose.Kill()
	client := newClient(compose.Containers["mockserver"])

	assert.Nil(t, client.ResetProxy())
}

func newClient(container *compose.Container) *mockserver.Client {
	client := mockserver.NewClient(
		fmt.Sprintf("http://%v:%v", compose.MustInferDockerHost(), container.MustGetFirstPublicPort(1080, "tcp")),
		fmt.Sprintf("http://%v:%v", compose.MustInferDockerHost(), container.MustGetFirstPublicPort(1090, "tcp")))

	compose.MustConnectWithDefaults(func() error {
		return client.ResetMocks()
	})

	return client
}

func newHttpClient(proxyURL string) *http.Client {
	parsedProxyURL, err := url.Parse(proxyURL)
	if err != nil {
		panic(err)
	}
	return &http.Client{
		Transport: &http.Transport{
			Proxy: http.ProxyURL(parsedProxyURL),
			TLSClientConfig: &tls.Config{
				InsecureSkipVerify: true,
			},
		},
	}
}
