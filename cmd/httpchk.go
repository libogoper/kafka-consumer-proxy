package main

import (
	"fmt"
	"net/http"
	"os"
	"runtime"
)

func main() {
	// 开启pprof
	http.HandleFunc("/httpchk", func(w http.ResponseWriter, r *http.Request) {
		output := fmt.Sprintf("runtime.GOMAXPROCS:%d", runtime.GOMAXPROCS(0))
		_, _ = w.Write([]byte(output))
	})
	ip := "0.0.0.0:80"
	if err := http.ListenAndServe(ip, nil); err != nil {
		fmt.Printf("start pprof failed on %s\n", ip)
		os.Exit(1)
	}
}
