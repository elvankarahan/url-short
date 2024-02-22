package main

import (
	"fmt"
	"github.com/lpernett/godotenv"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"log"
	"net/http"
	"os"
	"url-short/logger"
	"url-short/metrics"
	"url-short/url"
)

func setupRoutes(mux *http.ServeMux, API *url.API) {
	mux.HandleFunc("/", metrics.PrometheusHandler(API.Resolve))
	mux.HandleFunc("POST /api/v1", metrics.PrometheusHandler(API.Shorten))
	mux.Handle("/metrics", promhttp.Handler())

}
func main() {
	err := godotenv.Load()
	if err != nil {
		fmt.Println(err)
	}

	l := logger.New(false)

	API := url.New(l, 0)
	mux := http.NewServeMux()

	setupRoutes(mux, API)

	log.Fatal(http.ListenAndServe(os.Getenv("DOMAIN"), mux))
}
