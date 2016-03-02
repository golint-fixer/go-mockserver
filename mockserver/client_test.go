// +build integration

package mockserver_test

import (
	"testing"
	"os"
	"github.com/ory-am/dockertest"
	"time"
	"fmt"
	"net/http"
	"github.com/ibrt/go-mockserver/mockserver"
	"github.com/stretchr/testify/assert"
)

var (
	mockServerBaseUrl string
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
	c, ip, port, err := dockertest.SetupCustomContainer("jamesdbloom/mockserver", 1080, time.Minute)
	if err != nil {
		panic(err)
	}

	mockServerBaseUrl = fmt.Sprintf("http://%v:%v", ip, port)
	mockServerClient = mockserver.NewClient(mockServerBaseUrl)

	dockertest.ConnectToCustomContainer(mockServerBaseUrl, 10, time.Second, func (url string) bool {
		_, err := http.Get(url)
		return err == nil
	})
	return c
}

func TestMockAnyResponse(t *testing.T) {
	err := mockserver.NewMockAnyResponse().
		When(mockserver.NewRequest("GET", "/test")).
		Respond(mockserver.NewResponse(http.StatusOK)).
		Send(mockServerClient)
	assert.Nil(t, err)

	resp, err := http.Get(mockServerBaseUrl + "/test")
	assert.Nil(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)

	assert.Nil(t, mockServerClient.Reset())
	resp, err = http.Get(mockServerBaseUrl + "/test")
	assert.Nil(t, err)
	assert.Equal(t, http.StatusNotFound, resp.StatusCode)
}