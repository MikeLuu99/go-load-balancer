# Go Load Balancer

A simple, lightweight HTTP load balancer written in Go that distributes traffic across multiple backend servers using a round-robin algorithm.

## Features

- Round-robin load balancing
- Health checks for backend servers
- Simple API and deployment

## Installation

```bash
# Clone the repository
git clone https://github.com/go-load-balancer/go-load-balancer.git
cd go-load-balancer

# Build the load balancer and backend servers
make build
```

## Usage

### Start backend servers

Start multiple backend servers on different ports:

```bash
make run-backend-1  # Starts on port 8081
make run-backend-2  # Starts on port 8082
make run-backend-3  # Starts on port 8083
```

### Start the load balancer

```bash
make run-lb  # Starts on port 8080
```

The load balancer will distribute incoming requests to the three backend servers in a round-robin fashion.

## Configuration

The load balancer uses the following default configuration:
- Load balancer port: 8080
- Backend servers: localhost:8081, localhost:8082, localhost:8083
- Health check interval: 2 minutes

To modify the load balancer port:

```bash
./bin/loadbalancer -port 9090
```

To modify backend server ports:

```bash
./bin/backend -port 8084
```

## Project Structure

- `cmd/loadbalancer`: Load balancer implementation
- `cmd/backend`: Simple backend server for testing
- `pkg/loadbalancer`: Load balancing logic and health checking

## License

MIT