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

type CallbackFn func(map[string][]string)

func RunCallbackServer(wg *sync.WaitGroup, fn CallbackFn) {
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/" {
			http.Error(w, "404 not found.", http.StatusNotFound)
			return
		}

		_, err := w.Write([]byte(html))
		CheckErr(err)

		// Trigger the callback
		fn(r.URL.Query())

		wg.Done()
	})

	redirectUri := fmt.Sprintf("localhost:%s", viper.GetString("port"))

	l, err := net.Listen("tcp", redirectUri)
	CheckErr(err)

	// server is ready
	klog.V(50).InfoS(fmt.Sprintf("listening for response on: http://%s", redirectUri))

	err = http.Serve(l, nil)
	CheckErr(err)

}
