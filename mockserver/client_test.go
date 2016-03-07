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
)

var (
	mockServerMockBaseUrl string
	mockServerProxyBaseUrl string
	mockServerClient *mockserver.Client
)

func TestMain(m *testing.M) {
	os.Exit(testMain(m))
}

func testMain(m *testing.M) int {
	c := initMockServer()
	defer c.KillRemove()

	return m.Run()
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
	mockServerClient = mockserver.NewClient(mockServerMockBaseUrl, mockServerProxyBaseUrl)
	return c
}

func TestMockAnyResponse(t *testing.T) {
	err := mockserver.NewMockAnyResponse().
		When(mockserver.NewRequest("GET", "/test")).
		Respond(mockserver.NewResponse(http.StatusOK)).
		Send(mockServerClient)
	assert.Nil(t, err)

	resp, err := http.Get(mockServerMockBaseUrl + "/test")
	assert.Nil(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)

	assert.Nil(t, mockServerClient.MockReset())
	resp, err = http.Get(mockServerMockBaseUrl + "/test")
	assert.Nil(t, err)
	assert.Equal(t, http.StatusNotFound, resp.StatusCode)
}