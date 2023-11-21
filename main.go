package main

import (
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/joho/godotenv"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"

	"github.com/MrDan4es/regcloud_exporter/internal"
)

func main() {
	err := godotenv.Load(".env")

	if err != nil {
		log.Fatal("Error loading .env file")
	}

	port, ok := os.LookupEnv("REGCLOUD_HTTP_PORT")
	if !ok {
		log.Fatal("REGCLOUD_HTTP_PORT env not found")
	}

	token, ok := os.LookupEnv("REGCLOUD_TOKEN")
	if !ok {
		log.Fatal("REGCLOUD_TOKEN env not found")
	}

	req, err := internal.GetRequest(token)
	if err != nil {
		log.Fatal(err.Error())
	}

	httpClient := &http.Client{}

	collector := internal.NewBalanceCollector(httpClient, req)
	prometheus.MustRegister(collector)

	http.Handle("/metrics", promhttp.Handler())

	log.Printf("HTTP server starting on port %s...\n", port)
	log.Fatal(http.ListenAndServe(fmt.Sprintf(":%s", port), nil))
}
