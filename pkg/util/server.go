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

func CallbackUrl() string {
	return fmt.Sprintf("localhost:%s", viper.GetString("port"))
}

func callbackHandler(wg *sync.WaitGroup, r *http.Request, fn CallbackFn) {
	CheckErr(fn(r.URL.Query()))
	wg.Done()
}

type CallbackFn func(map[string][]string) error

func RunCallbackServer(wg *sync.WaitGroup, fn CallbackFn) {
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/" {
			http.Error(w, "404 not found.", http.StatusNotFound)
			return
		}

		_, err := w.Write([]byte(html))
		CheckErr(err)

		wg.Add(1)
		go callbackHandler(wg, r, fn)

		wg.Done()
	})

	l, err := net.Listen("tcp", CallbackUrl())
	CheckErr(err)

	// server is ready
	klog.V(50).InfoS(fmt.Sprintf("listening for response on: http://%s", CallbackUrl()))
	wg.Done()

	err = http.Serve(l, nil)
	CheckErr(err)

}
