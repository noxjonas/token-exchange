package util

import (
	"fmt"
	"k8s.io/klog/v2"
	"log"
	"os/exec"
	"runtime"
	"sync"
)

func OpenBrowser(wg *sync.WaitGroup, url string) {
	var err error

	switch runtime.GOOS {
	case "linux":
		err = exec.Command("xdg-open", url).Start()
	case "windows":
		err = exec.Command("rundll32", "url.dll,FileProtocolHandler", url).Start()
	case "darwin":
		err = exec.Command("open", url).Start()
	default:
		err = fmt.Errorf("unsupported platform")
	}
	if err == nil {
		klog.V(50).InfoS(fmt.Sprintf("browser oppened at: %s", url))
		wg.Done()
	}
	if err != nil {
		log.Fatal(err)
	}

}
