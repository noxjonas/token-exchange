package util

import (
	_ "embed"
	"fmt"
	"github.com/spf13/viper"
	"k8s.io/klog/v2"
	"net"
	"net/http"
	"sync"
)

//go:embed static/index.html
var html string

func GetUrl() string {
	port := viper.GetString("port")
	if port == "" {
		port = "8080"
		klog.V(10).InfoS("callback.GetUrl: no port found in viper. defaulting to 8080")
	}
	return fmt.Sprintf("localhost:%s", port)
}

// must run in a separate goroutine to ensure that callback HTML is served before the program ends
func handler(wg *sync.WaitGroup, r *http.Request, fn Fn) {
	CheckErr(fn(r.URL.Query()))
	wg.Done()
}

type Fn func(map[string][]string) error

// RunServer emits wg.Done() when server starts listening, and once when callback is triggered
func RunServer(wg *sync.WaitGroup, fn Fn) {

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/" {
			http.Error(w, "404 not found.", http.StatusNotFound)
			return
		}

		_, err := w.Write([]byte(html))
		CheckErr(err)

		wg.Add(1)
		go handler(wg, r, fn)

		wg.Done()
	})

	l, err := net.Listen("tcp", GetUrl())
	CheckErr(err)

	// server is ready
	klog.V(50).InfoS(fmt.Sprintf("listening for response on: http://%s", GetUrl()))
	wg.Done()

	err = http.Serve(l, nil)
	CheckErr(err)

}
