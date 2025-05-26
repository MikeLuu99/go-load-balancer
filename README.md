# Go Load Balancer

A simple, robust HTTP load balancer implementation in Go.

## Features

- Round-robin load balancing algorithm
- Automatic health checks for backend servers
- Graceful handling of failed backend servers
- Reverse proxy functionality with custom headers
- Configurable health check intervals

## Getting Started

### Prerequisites

- Go 1.24 or higher

### Running the Load Balancer

```bash
cd go-load-balancer
go run loadbalancer/load-balancer.go [options]
```

Options:
- `-port`: Port to serve on (default: 8080)
- `-check-interval`: Interval for health checking backends (default: 1m)

### Running Backend Servers

For testing, you can start multiple instances of the simple backend server:

```bash
# Start three backends on different ports
go run backend/simple-backend.go -port 8081
go run backend/simple-backend.go -port 8082
go run backend/simple-backend.go -port 8083
```

## How It Works

1. The load balancer distributes incoming HTTP requests to backend servers in a round-robin fashion
2. Periodic health checks ensure that requests are only sent to healthy backends
3. If a backend fails, it's marked as unavailable until it becomes healthy again
4. All requests are proxied transparently with added headers for tracking

## Project Structure

- `loadbalancer/`: Contains the load balancer implementation
- `backend/`: Contains a simple backend server for testing
