package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

func main() {
	if len(os.Args) != 2 {
		fmt.Fprintln(os.Stderr, "Usage", os.Args[0], "<config_file>")
		os.Exit(1)
	}

	client := &http.Client{
		Timeout: time.Second * 10,
	}
	config := NewConfig(os.Args[1])
	collector := NewCollector(&config, client)

	prometheus.MustRegister(&collector)
	http.Handle("/metrics", promhttp.Handler())
	log.Println(http.ListenAndServe(config.Listen, nil))
}
