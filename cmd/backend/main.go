package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"
)

func main() {
	port := flag.Int("port", 8080, "Port to serve on")
	flag.Parse()

	mux := http.NewServeMux()
	// Handler for the root path
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		// Get hostname for identification
		hostname, err := os.Hostname()
		if err != nil {
			hostname = "unknown"
		}

		// Add a short delay to simulate processing time (optional)
		time.Sleep(100 * time.Millisecond)

		// Log the request
		log.Printf("Backend %d received request: %s %s", *port, r.Method, r.URL.Path)

		// Return a response that identifies this backend
		fmt.Fprintf(w, "Backend server on port %d\n", *port)
		fmt.Fprintf(w, "Host: %s\n", hostname)
		fmt.Fprintf(w, "Request path: %s\n", r.URL.Path)
		fmt.Fprintf(w, "Request method: %s\n", r.Method)
		fmt.Fprintf(w, "Request headers:\n")

		// Print all request headers
		for name, values := range r.Header {
			for _, value := range values {
				fmt.Fprintf(w, "  %s: %s\n", name, value)
			}
		}

		mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
			fmt.Fprintf(w, "healthy")
		})

		server := http.Server{
			Addr:    fmt.Sprintf(":%d", *port),
			Handler: mux,
		}

		if err := server.ListenAndServe(); err != nil {
			log.Fatal(err)
		}

	})
}
