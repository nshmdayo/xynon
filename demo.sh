#!/bin/bash

# nproxy Demo Script
# This script runs a demo of the proxy and mock server

set -e

# Functions for colored output
print_header() {
    echo -e "\n\033[1;36m=== $1 ===\033[0m"
}

print_info() {
    echo -e "\033[1;32m[INFO]\033[0m $1"
}

print_warning() {
    echo -e "\033[1;33m[WARNING]\033[0m $1"
}

print_error() {
    echo -e "\033[1;31m[ERROR]\033[0m $1"
}

# Function to clean up processes
cleanup() {
    print_info "Cleaning up processes..."
    if [ ! -z "$MOCK_PID" ]; then
        kill $MOCK_PID 2>/dev/null || true
        print_info "Mock server stopped (PID: $MOCK_PID)"
    fi
    if [ ! -z "$PROXY_PID" ]; then
        kill $PROXY_PID 2>/dev/null || true
        print_info "Proxy server stopped (PID: $PROXY_PID)"
    fi
}

# Execute cleanup on Ctrl+C
trap cleanup EXIT

# Check if Go is installed
if ! command -v go &> /dev/null; then
    print_error "Go is not installed. Please install Go first."
    exit 1
fi

print_header "nproxy + Mock Server Demo"

# Select demo type based on arguments
DEMO_TYPE=${1:-"basic"}

case $DEMO_TYPE in
    "basic")
        print_info "Starting Basic Proxy Demo"
        PROXY_CMD="go run app/main.go -addr :8080"
        ;;
    "mitm")
        print_info "Starting MITM Proxy Demo"
        PROXY_CMD="go run app/main.go -mitm -addr :8080"
        ;;
    "mitm-modify")
        print_info "Starting MITM Proxy Demo with Modification"
        PROXY_CMD="go run app/main.go -mitm -modify -v -addr :8080"
        ;;
    *)
        print_error "Unknown demo type: $DEMO_TYPE"
        echo "Usage: $0 [basic|mitm|mitm-modify]"
        exit 1
        ;;
esac

# 1. Start mock server
print_header "Starting Mock Server"
print_info "Starting mock server on :9090..."
go run app/main.go -mock -addr :9090 &
MOCK_PID=$!
print_info "Mock server started (PID: $MOCK_PID)"

# Wait a bit for mock server to start
sleep 2

# Wait a bit for mock server to start
sleep 2

# Check mock server operation
print_info "Testing mock server directly..."
if curl -s http://localhost:9090/health > /dev/null; then
    print_info "✓ Mock server is responding"
else
    print_error "✗ Mock server is not responding"
    exit 1
fi

# 2. Start proxy server
print_header "Starting Proxy Server"
print_info "Starting proxy server on :8080..."
$PROXY_CMD &
PROXY_PID=$!
print_info "Proxy server started (PID: $PROXY_PID)"

# Wait a bit for proxy server to start
sleep 2

# 3. Run tests
print_header "Running Tests"

print_info "Test 1: Health check via proxy"
echo "Command: curl -x localhost:8080 http://localhost:9090/health"
curl -x localhost:8080 http://localhost:9090/health | jq . || curl -x localhost:8080 http://localhost:9090/health
echo ""

print_info "Test 2: Users API via proxy"
echo "Command: curl -x localhost:8080 http://localhost:9090/api/users"
curl -x localhost:8080 http://localhost:9090/api/users | jq . || curl -x localhost:8080 http://localhost:9090/api/users
echo ""

print_info "Test 3: Echo API via proxy"
echo "Command: curl -x localhost:8080 -X POST -H 'Content-Type: application/json' -d '{\"test\":\"data\"}' http://localhost:9090/api/echo"
curl -x localhost:8080 -X POST -H "Content-Type: application/json" -d '{"test":"data"}' http://localhost:9090/api/echo | jq . || curl -x localhost:8080 -X POST -H "Content-Type: application/json" -d '{"test":"data"}' http://localhost:9090/api/echo
echo ""

print_info "Test 4: Default endpoint via proxy"
echo "Command: curl -x localhost:8080 http://localhost:9090/unknown"
curl -x localhost:8080 http://localhost:9090/unknown | jq . || curl -x localhost:8080 http://localhost:9090/unknown
echo ""

print_header "Demo Complete"
print_info "All tests completed successfully!"
print_info "Check the server logs above to see how requests are processed through the proxy."

if [[ $DEMO_TYPE == "mitm-modify" ]]; then
    print_info "Notice: With MITM modification enabled, you should see additional headers being added to requests and responses."
fi

print_warning "Press Ctrl+C to stop the servers and exit."

# Keep servers running
wait
