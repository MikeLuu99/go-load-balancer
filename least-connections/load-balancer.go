package main

import (
	"net/http/httputil"
	"net/url"
	"sync"
)

type Backend struct {
	URL                 *url.URL
	mux                 sync.RWMutex
	NumberOfConnections int
	ReverseProxy        *httputil.ReverseProxy
}

func (b *Backend) SetNumberOfConnections(numberOfConnections int) {
	b.mux.Lock()
	defer b.mux.Unlock()
	b.NumberOfConnections = numberOfConnections
}

func (b *Backend) IncrementNumberOfConnections() {
	b.mux.Lock()
	defer b.mux.Unlock()
	b.NumberOfConnections++
}

type LoadBalancer struct {
	backends []*Backend
	current  uint64
}

func (lb *LoadBalancer) FindBackend() *Backend {
	if len(lb.backends) == 0 {
		return nil
	}

	var leastConnectionBackend int
	for i, backend := range lb.backends {
		if backend.NumberOfConnections < lb.backends[leastConnectionBackend].NumberOfConnections {
			leastConnectionBackend = i
		}
	}
	return lb.backends[leastConnectionBackend]
}
