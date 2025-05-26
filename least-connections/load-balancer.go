package main

import (
	"flag"
	"fmt"
	"log"
	"net"
	"net/http"
	"net/http/httputil"
	"net/url"
	"sync"
	"time"
)

// Backend represents a backend server
type Backend struct {
	URL                 *url.URL
	Alive               bool
	mux                 sync.RWMutex
	NumberOfConnections int
	ReverseProxy        *httputil.ReverseProxy
}

// SetAlive updates the alive status of backend
func (b *Backend) SetAlive(alive bool) {
	b.mux.Lock()
	defer b.mux.Unlock()
	b.Alive = alive
}

// IsAlive returns true when backend is alive
func (b *Backend) IsAlive() (alive bool) {
	b.mux.RLock()
	defer b.mux.RUnlock()
	alive = b.Alive
	return
}

// SetNumberOfConnections sets the number of connections for the backend
func (b *Backend) SetNumberOfConnections(numberOfConnections int) {
	b.mux.Lock()
	defer b.mux.Unlock()
	b.NumberOfConnections = numberOfConnections
}

// IncrementNumberOfConnections increments the connection count
func (b *Backend) IncrementNumberOfConnections() {
	b.mux.Lock()
	defer b.mux.Unlock()
	b.NumberOfConnections++
}

// DecrementNumberOfConnections decrements the connection count
func (b *Backend) DecrementNumberOfConnections() {
	b.mux.Lock()
	defer b.mux.Unlock()
	if b.NumberOfConnections > 0 {
		b.NumberOfConnections--
	}
}

// GetNumberOfConnections returns the current connection count
func (b *Backend) GetNumberOfConnections() int {
	b.mux.RLock()
	defer b.mux.RUnlock()
	return b.NumberOfConnections
}

// LoadBalancer represents a load balancer using least connections algorithm
type LoadBalancer struct {
	backends []*Backend
	current  uint64 // kept for compatibility, not used in least connections
}

// FindBackend returns the backend with the least number of active connections
func (lb *LoadBalancer) FindBackend() *Backend {
	if len(lb.backends) == 0 {
		return nil
	}

	var leastConnectionBackend *Backend
	minConnections := -1

	// Find the alive backend with the least connections
	for _, backend := range lb.backends {
		if !backend.IsAlive() {
			continue
		}

		connections := backend.GetNumberOfConnections()
		if minConnections == -1 || connections < minConnections {
			minConnections = connections
			leastConnectionBackend = backend
		}
	}

	return leastConnectionBackend
}

// isBackendAlive checks whether a backend is alive by establishing a TCP connection
func isBackendAlive(u *url.URL) bool {
	timeout := 2 * time.Second
	conn, err := net.DialTimeout("tcp", u.Host, timeout)
	if err != nil {
		log.Printf("Site unreachable: %s", err)
		return false
	}
	defer conn.Close()
	return true
}

// HealthCheck pings the backends and updates their status
func (lb *LoadBalancer) HealthCheck() {
	for _, b := range lb.backends {
		status := isBackendAlive(b.URL)
		b.SetAlive(status)
		if status {
			log.Printf("Backend %s is alive (connections: %d)", b.URL, b.GetNumberOfConnections())
		} else {
			log.Printf("Backend %s is dead", b.URL)
		}
	}
}

// HealthCheckPeriodically runs a routine health check every interval
func (lb *LoadBalancer) HealthCheckPeriodically(interval time.Duration) {
	t := time.NewTicker(interval)
	defer t.Stop()
	for {
		select {
		case <-t.C:
			lb.HealthCheck()
		}
	}
}

// ServeHTTP implements the http.Handler interface for the LoadBalancer
func (lb *LoadBalancer) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	backend := lb.FindBackend()
	if backend == nil {
		http.Error(w, "Service Unavailable", http.StatusServiceUnavailable)
		return
	}

	// Increment connection count before handling request
	backend.IncrementNumberOfConnections()

	// Log the request
	log.Printf("Routing request %s %s to backend %s (connections: %d)",
		r.Method, r.URL.Path, backend.URL, backend.GetNumberOfConnections())

	// Create a custom response writer to track when the request completes
	wrappedWriter := &responseWriter{
		ResponseWriter: w,
		backend:        backend,
	}

	// Forward the request to the backend
	backend.ReverseProxy.ServeHTTP(wrappedWriter, r)
}

// responseWriter wraps http.ResponseWriter to decrement connection count when done
type responseWriter struct {
	http.ResponseWriter
	backend *Backend
	written bool
}

func (rw *responseWriter) Write(data []byte) (int, error) {
	if !rw.written {
		rw.written = true
		// Decrement connection count when response starts being written
		defer rw.backend.DecrementNumberOfConnections()
	}
	return rw.ResponseWriter.Write(data)
}

func (rw *responseWriter) WriteHeader(statusCode int) {
	if !rw.written {
		rw.written = true
		// Decrement connection count when response header is written
		defer rw.backend.DecrementNumberOfConnections()
	}
	rw.ResponseWriter.WriteHeader(statusCode)
}

func main() {
	// Parse command line flags
	port := flag.Int("port", 8080, "Port to serve on")
	checkInterval := flag.Duration("check-interval", time.Minute, "Interval for health checking backends")
	flag.Parse()

	// Configure backends (in a real application, this might come from a config file)
	serverList := []string{
		"http://localhost:8081",
		"http://localhost:8082",
		"http://localhost:8083",
	}

	// Create load balancer
	lb := LoadBalancer{}

	// Initialize backends
	for _, serverURL := range serverList {
		url, err := url.Parse(serverURL)
		if err != nil {
			log.Fatal(err)
		}

		proxy := httputil.NewSingleHostReverseProxy(url)

		// Customize the reverse proxy director
		originalDirector := proxy.Director
		proxy.Director = func(r *http.Request) {
			originalDirector(r)
			r.Header.Set("X-Proxy", "Least-Connections-Load-Balancer")
		}

		// Add custom error handler
		proxy.ErrorHandler = func(w http.ResponseWriter, r *http.Request, err error) {
			log.Printf("Proxy error: %v", err)
			http.Error(w, "Service Unavailable", http.StatusServiceUnavailable)

			// Mark the backend as down and decrement connection count
			for _, b := range lb.backends {
				if b.URL.String() == url.String() {
					b.SetAlive(false)
					b.DecrementNumberOfConnections()
					break
				}
			}
		}

		lb.backends = append(lb.backends, &Backend{
			URL:                 url,
			Alive:               true,
			NumberOfConnections: 0,
			ReverseProxy:        proxy,
		})
		log.Printf("Configured backend: %s", url)
	}

	// Initial health check
	lb.HealthCheck()

	// Start periodic health check
	go lb.HealthCheckPeriodically(*checkInterval)

	// Start the server
	server := http.Server{
		Addr:    fmt.Sprintf(":%d", *port),
		Handler: &lb,
		// Set reasonable timeouts
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 10 * time.Second,
		IdleTimeout:  120 * time.Second,
	}

	defer server.Close()

	log.Printf("Least Connections Load Balancer started at :%d\n", *port)
	if err := server.ListenAndServe(); err != nil {
		log.Fatal(err)
	}
}
