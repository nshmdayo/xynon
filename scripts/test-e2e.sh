#!/bin/bash
set -e

echo "==> Building proxy..."
make build

echo "==> Building plugins..."
make wasm

echo "==> Starting local echo server..."
go run scripts/echo.go &
ECHO_PID=$!

echo "==> Starting proxy..."
./bin/xynon -config examples/config.yaml &
PROXY_PID=$!

# Ensure processes are killed on exit
trap 'kill $PROXY_PID $ECHO_PID 2>/dev/null' EXIT

# Wait for proxy and echo server to start
echo "==> Waiting for services to start..."
timeout 60 bash -c 'until curl -s -f http://localhost:8081 >/dev/null; do sleep 1; done'
# Since the proxy doesn't serve requests itself without forwarding, wait until the proxy port is open using a dummy curl
timeout 60 bash -c 'until curl -s http://localhost:8080 >/dev/null; do sleep 1; done' || true
sleep 1

echo "==> Running E2E tests..."

# Test 1: auth plugin without token (expect 401)
RESPONSE=$(curl -s -D - -x http://localhost:8080 http://localhost:8081)
if echo "$RESPONSE" | grep -q "HTTP/1.1 401 Unauthorized"; then
    echo "✅ auth plugin (no token) test passed"
else
    echo "❌ auth plugin (no token) test failed"
    echo "$RESPONSE"
    exit 1
fi

# Test 2: auth plugin with token (expect 200) + add-header plugin
export JWT_TOKEN=$(go run scripts/gen_jwt.go)
RESPONSE=$(curl -s -D - -H "Authorization: Bearer $JWT_TOKEN" -x http://localhost:8080 http://localhost:8081)
if echo "$RESPONSE" | grep -q "X-Xynon: true"; then
    echo "✅ auth plugin (with token) & add-header plugin test passed"
else
    echo "❌ auth plugin (with token) & add-header plugin test failed"
    echo "$RESPONSE"
    exit 1
fi

# Test 3: CONNECT request without token (expect 401)
RESPONSE=$(python3 -c '
import socket
s = socket.socket()
s.connect(("localhost", 8080))
s.sendall(b"CONNECT localhost:8081 HTTP/1.1\r\nHost: localhost:8081\r\n\r\n")
print(s.recv(1024).decode("utf-8"))
s.close()
')
if echo "$RESPONSE" | grep -q "HTTP/1.1 401 Unauthorized"; then
    echo "✅ CONNECT auth plugin (no token) test passed"
else
    echo "❌ CONNECT auth plugin (no token) test failed"
    echo "$RESPONSE"
    exit 1
fi

# Test 4: CONNECT request with token (expect 200)
RESPONSE=$(python3 -c '
import socket, os
s = socket.socket()
s.connect(("localhost", 8080))
token = os.environ.get("JWT_TOKEN", "")
req = "CONNECT localhost:8081 HTTP/1.1\r\nHost: localhost:8081\r\nAuthorization: Bearer " + token + "\r\n\r\n"
s.sendall(req.encode("utf-8"))
print(s.recv(1024).decode("utf-8"))
s.close()
')
if echo "$RESPONSE" | grep -q "HTTP/1.1 200 Connection established"; then
    echo "✅ CONNECT auth plugin (with token) test passed"
else
    echo "❌ CONNECT auth plugin (with token) test failed"
    echo "$RESPONSE"
    exit 1
fi


# Test 5: Caching Plugin
export JWT_TOKEN=$(go run scripts/gen_jwt.go)
# Request 1
RESP1=$(curl -s -H "Authorization: Bearer $JWT_TOKEN" -H "X-Rand: 111" -x http://localhost:8080 http://localhost:8081/cache-test)
# Request 2 (should be cached, so X-Rand should be 111 in the body, even if we send 222)
RESP2=$(curl -s -H "Authorization: Bearer $JWT_TOKEN" -H "X-Rand: 222" -x http://localhost:8080 http://localhost:8081/cache-test)

if echo "$RESP1" | grep -q "X-Rand: 111" && echo "$RESP2" | grep -q "X-Rand: 111"; then
    echo "✅ caching plugin test passed"
else
    echo "❌ caching plugin test failed"
    echo "RESP1: $RESP1"
    echo "RESP2: $RESP2"
    exit 1
fi

echo "==> All E2E tests passed!"

