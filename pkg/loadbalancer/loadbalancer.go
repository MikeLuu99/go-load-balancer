package loadbalancer

import (
	"log"
	"net"
	"net/http"
	"net/http/httputil"
	"net/url"
	"sync"
	"sync/atomic"
	"time"
)

type Backend struct {
	URL          *url.URL
	Alive        bool
	mux          sync.RWMutex
	ReverseProxy *httputil.ReverseProxy
}

type LoadBalancer struct {
	backends []*Backend
	current  uint64
}

func (b *Backend) SetAlive(alive bool) {
	b.mux.Lock()
	b.Alive = alive
	b.mux.Unlock()
}

func (b *Backend) IsAlive() (alive bool) {
	b.mux.RLock()
	alive = b.Alive
	b.mux.RUnlock()
	return
}

func (lb *LoadBalancer) NextBackend() *Backend {
	next := atomic.AddUint64(&lb.current, uint64(1)) % uint64(len(lb.backends))

	for i := range lb.backends {
		idx := (int(next) + i) % len(lb.backends)
		if lb.backends[idx].IsAlive() {
			return lb.backends[idx]
		}
	}

	return nil
}

func IsBackendAlive(u *url.URL) bool {
	timeout := 2 * time.Second
	conn, err := net.DialTimeout("tcp", u.Host, timeout)
	if err != nil {
		log.Printf("Site unreachable: %s", err)
		return false
	}
	defer conn.Close()
	return true
}

func (lb *LoadBalancer) HealthCheck() {
	for _, backend := range lb.backends {
		status := IsBackendAlive(backend.URL)
		backend.SetAlive(status)
		if status {
			log.Printf("Backend %s is alive", backend.URL)
		} else {
			log.Printf("Backend %s is dead", backend.URL)
		}
	}
}

func (lb *LoadBalancer) HealthCheckPeriodically(duration time.Duration) {
	t := time.NewTicker(duration)
	for {
		select {
		case <-t.C:
			lb.HealthCheck()
		}
	}
}

func (lb *LoadBalancer) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	backend := lb.NextBackend()
	if backend == nil {
		http.Error(w, "Service Unavailable", http.StatusServiceUnavailable)
		return
	}
	backend.ReverseProxy.ServeHTTP(w, r)
}

func NewLoadBalancer(serverURLs []string) (*LoadBalancer, error) {
	var backends []*Backend
	
	for _, serverURL := range serverURLs {
		url, err := url.Parse(serverURL)
		if err != nil {
			return nil, err
		}

		proxy := httputil.NewSingleHostReverseProxy(url)
		proxy.ErrorHandler = func(w http.ResponseWriter, r *http.Request, err error) {
			log.Printf("Error: %v", err)
			http.Error(w, "Service Unavailable", http.StatusServiceUnavailable)
		}

		backends = append(backends, &Backend{
			URL:          url,
			Alive:        true,
			ReverseProxy: proxy,
		})
	}
	
	return &LoadBalancer{
		backends: backends,
	}, nil
}