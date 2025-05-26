package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/go-load-balancer/pkg/loadbalancer"
)

func main() {
	port := flag.Int("port", 8080, "Port to serve on")
	flag.Parse()

	serverList := []string{
		"http://localhost:8081",
		"http://localhost:8082",
		"http://localhost:8083",
	}

	lb, err := loadbalancer.NewLoadBalancer(serverList)
	if err != nil {
		log.Fatal(err)
	}

	// Start health checker in background
	go lb.HealthCheckPeriodically(2 * time.Minute)

	server := http.Server{
		Addr:    fmt.Sprintf(":%d", *port),
		Handler: lb,
	}

	log.Printf("Load balancer serving on port %d", *port)
	if err := server.ListenAndServe(); err != nil {
		log.Fatal(err)
	}
}