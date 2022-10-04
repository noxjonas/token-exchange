package util

import (
	"github.com/spf13/viper"
	"net/http"
	"sync"
	"testing"
)

func setPort(port int) {
	// port is pulled from viper
	viper.Set("port", port)
}

func TestGetUrl(t *testing.T) {
	url := GetUrl()
	if url != "localhost:8080" {
		t.Error("GetUrl: without viper config, default callback server should be localhost:8080")
	}

	setPort(8081)
	url = GetUrl()
	if url != "localhost:8081" {
		t.Error("GetUrl: failed to get port from viper")
	}
}

func TestRunServer(t *testing.T) {
	setPort(8090)

	wg := new(sync.WaitGroup)
	wg.Add(1)
	go RunServer(wg, testCallbackFn)
	wg.Wait() // wait for callback server to be ready

	wg.Add(1)

	get, err := http.Get("http://localhost:8090/?code=123&other=abc")
	if err != nil {
		t.Errorf("callback test get request failed: err=%s", err)
	}
	if get.StatusCode != http.StatusOK {
		t.Errorf("callback did not respond with 200: status=%d", get.StatusCode)
	}

	if resultParams.Code != "123" || resultParams.Other != "abc" {
		t.Errorf("invalid parameters returned to callback: code=%s, other=%s", resultParams.Code, resultParams.Other)
	}

	wg.Wait()
}

type resultParamsT struct {
	Code  string
	Other string
}

var resultParams = resultParamsT{}

func testCallbackFn(params map[string][]string) error {
	resultParams.Code = params["code"][0]
	resultParams.Other = params["other"][0]
	return nil
}
