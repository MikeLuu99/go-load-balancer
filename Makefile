# Build commands
build-lb:
	go build -o bin/loadbalancer ./cmd/loadbalancer

build-backend:
	go build -o bin/backend ./cmd/backend

build: build-lb build-backend

# Run commands
run-lb:
	./bin/loadbalancer

run-backend-1:
	./bin/backend -port 8081

run-backend-2:
	./bin/backend -port 8082

run-backend-3:
	./bin/backend -port 8083

# Clean command
clean:
	rm -rf bin/

# Create bin directory
init:
	mkdir -p bin