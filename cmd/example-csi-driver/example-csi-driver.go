package main

import (
	"math/rand"
	"os"
	"runtime"
	"time"

	"github.com/wangweihong/example-csi-driver/internal/csidriver"
)

func main() {
	rand.Seed(time.Now().UTC().UnixNano())
	if len(os.Getenv("GOMAXPROCS")) == 0 {
		runtime.GOMAXPROCS(runtime.NumCPU())
	}

	csidriver.NewApp("example-csi-driver").Run()
}
