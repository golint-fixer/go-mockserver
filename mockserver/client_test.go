// +build integration

package mockserver_test

import (
	"testing"
	"os"
	"github.com/ibrt/dockertest"
	"time"
	"fmt"
	"net/http"
	"github.com/ibrt/go-mockserver/mockserver"
	"github.com/stretchr/testify/assert"
	"net/url"
	"crypto/tls"
	"bytes"
)

var (
	mockServerMockBaseUrl string
	mockServerProxyBaseUrl string
	mockServerClient *mockserver.Client
	proxyClient *http.Client
)

func TestMain(m *testing.M) {
	os.Exit(testMain(m))
}

func testMain(m *testing.M) (result int) {
	id := initMockServer()
	defer func() {
		if result == 0 {
			id.KillRemove()
		} else {
			id.Kill()
		}
	}()

	result = m.Run()
	return
}

func initMockServer() dockertest.ContainerID {
	c, err := dockertest.ConnectToMockserver(15, time.Millisecond*500,
		func(url string) bool {
			req, err := http.NewRequest("PUT", fmt.Sprintf("%v/reset", url), nil)
			if err != nil {
				return false
			}
			_, err = http.DefaultClient.Do(req)
			if err == nil {
				mockServerMockBaseUrl = url
			}
			return err == nil
		},
		func(url string) bool {
			req, err := http.NewRequest("PUT", fmt.Sprintf("%v/reset", url), nil)
			if err != nil {
				return false
			}
			_, err = http.DefaultClient.Do(req)
			if err == nil {
				mockServerProxyBaseUrl = url
			}
			return err == nil
		})
	if err != nil {
		panic(err)
	}
	proxyUrl, err := url.Parse(mockServerProxyBaseUrl)
	if err != nil {
		panic(err)
	}
	proxyClient = &http.Client{
		Transport: &http.Transport{
			Proxy: http.ProxyURL(proxyUrl),
			TLSClientConfig: &tls.Config{
				InsecureSkipVerify: true,
			},
		},
	}
	mockServerClient = mockserver.NewClient(mockServerMockBaseUrl, mockServerProxyBaseUrl)
	return c
}

func TestMockAnyResponse(t *testing.T) {
	err := mockServerClient.MockAnyResponse(
		mockserver.NewMockAnyResponse().
			When(mockserver.NewRequest("GET", "/test")).
			Respond(mockserver.NewResponse(http.StatusOK)))
	assert.Nil(t, err)

	resp, err := http.Get(mockServerMockBaseUrl + "/test")
	assert.Nil(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	t.Fail()
}

func TestMockReset(t *testing.T) {
	assert.Nil(t, mockServerClient.ResetMocks())
}

func TestProxy(t *testing.T) {
	_, err := proxyClient.Get("https://www.google.com/")
	assert.Nil(t, err)

	_, err = proxyClient.Post("http://www.google.com/", "application/json", bytes.NewBuffer([]byte("{ \"hello\": true }")))
	assert.Nil(t, err)

	err = mockServerClient.VerifyProxy(
		mockserver.NewVerify().
			MatchRequest(mockserver.NewRequest("GET", "/")).
			WithTimes(1, true))
	assert.Nil(t, err)

	_, err = mockServerClient.RetrieveProxy(
		mockserver.NewRetrieve().
			MatchRequest(mockserver.NewRequest("GET", "/")))
	assert.Nil(t, err)

	_, err = mockServerClient.RetrieveProxy(
		mockserver.NewRetrieve().
		MatchRequest(mockserver.NewRequest("POST", "/")))
	assert.Nil(t, err)
}

func TestProxyReset(t *testing.T) {
	assert.Nil(t, mockServerClient.ResetProxy())
}