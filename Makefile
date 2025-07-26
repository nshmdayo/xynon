start:
	docker build -t nproxy -f ./build/Dockerfile .
	docker run --name nproxy -p 8000:8000 -it nproxy

mitm:
	docker build -t nproxy -f ./build/Dockerfile .
	docker run --name nproxy-mitm -p 8080:8080 -v ./certs:/app/certs -it nproxy -mitm -addr :8080

mitm-modify:
	docker build -t nproxy -f ./build/Dockerfile .
	docker run --name nproxy-mitm -p 8080:8080 -v ./certs:/app/certs -it nproxy -mitm -modify -v -addr :8080

stop:
	docker stop nproxy
	docker rm nproxy
	docker rmi nproxy

stop-mitm:
	docker stop nproxy-mitm || true
	docker rm nproxy-mitm || true

build:
	go build -o bin/nproxy app/main.go

run:
	go run app/main.go

run-mitm:
	go run app/main.go -mitm -addr :8080

run-mitm-modify:
	go run app/main.go -mitm -modify -v -addr :8080

# Mock server commands
run-mock:
	go run app/main.go -mock -addr :9090

# Test proxy with mock server (run these in separate terminals)
test-proxy-mock:
	@echo "Starting mock server on :9090 and proxy on :8080"
	@echo "Run 'make run-mock' in one terminal and 'make run' in another terminal"
	@echo "Then test with: curl -x localhost:8080 http://localhost:9090/health"

# Quick test commands for proxy + mock
demo-basic:
	@echo "Testing basic proxy with mock server..."
	@echo "1. Start mock server: make run-mock (in another terminal)"
	@echo "2. Start proxy: make run (in another terminal)"
	@echo "3. Test: curl -x localhost:8080 http://localhost:9090/health"

demo-mitm:
	@echo "Testing MITM proxy with mock server..."
	@echo "1. Start mock server: make run-mock (in another terminal)"
	@echo "2. Start MITM proxy: make run-mitm-modify (in another terminal)"
	@echo "3. Test: curl -x localhost:8080 http://localhost:9090/health"

test:
	go test ./app/proxy/ ./app/mock/

test-verbose:
	go test -v ./app/proxy/ ./app/mock/

test-short:
	go test ./app/proxy/ -short

test-bench:
	go test ./app/proxy/ -bench=.

# Build commands
build-linux:
	GOOS=linux GOARCH=amd64 go build -o bin/nproxy-linux app/main.go

build-windows:
	GOOS=windows GOARCH=amd64 go build -o bin/nproxy.exe app/main.go

build-all: build build-linux build-windows

# Help command
help:
	@echo "Available commands:"
	@echo "  start           - Start basic proxy with Docker"
	@echo "  stop            - Stop and cleanup basic proxy"
	@echo "  restart         - Restart basic proxy"
	@echo "  run             - Run basic proxy locally"
	@echo "  run-mitm        - Run MITM proxy locally"
	@echo "  run-mitm-modify - Run MITM proxy with request/response modification"
	@echo "  mitm            - Start MITM proxy with Docker"
	@echo "  mitm-modify     - Start MITM proxy with modification enabled"
	@echo "  stop-mitm       - Stop and cleanup MITM proxy"
	@echo "  run-mock        - Run mock server locally"
	@echo "  demo-basic      - Instructions for testing basic proxy with mock"
	@echo "  demo-mitm       - Instructions for testing MITM proxy with mock"
	@echo "  test-proxy-mock - Instructions for testing proxy with mock server"
	@echo "  test-verbose    - Run tests with verbose output"
	@echo "  test-short      - Run short tests only"
	@echo "  test-bench      - Run benchmark tests"
	@echo "  build           - Build binary"
	@echo "  build-all       - Build for all platforms"
	@echo "  clean           - Clean build artifacts and certificates"
	@echo "  help            - Show this help message"

.PHONY: start stop restart run run-mitm run-mitm-modify test test-verbose test-short test-bench build build-linux build-windows build-all clean mitm mitm-modify stop-mitm help

clean: stop stop-mitm
	docker rmi nproxy || true
	rm -f bin/nproxy
	rm -rf certs/

restart: stop start