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

# Test 5: Rate Limiter
echo "==> Running Test 5: Rate Limiter"
export JWT_TOKEN=$(go run scripts/gen_jwt.go)
# Use a unique X-Forwarded-For for this test so we have a fresh bucket.
IP="10.0.0.5"
for idx in 1 2 3 4 5; do
    STATUS=$(curl -s -o /dev/null -w "%{http_code}" -H "Authorization: Bearer $JWT_TOKEN" -H "X-Forwarded-For: $IP" -x http://localhost:8080 http://localhost:8081)
    if [ "$STATUS" != "200" ]; then
        echo "❌ rate limiter test failed (request $idx returned $STATUS, expected 200)"
        exit 1
    fi
done

# 6th request should fail with 429
RESPONSE=$(curl -s -D - -H "Authorization: Bearer $JWT_TOKEN" -H "X-Forwarded-For: $IP" -x http://localhost:8080 http://localhost:8081)
if echo "$RESPONSE" | grep -q "HTTP/1.1 429 Too Many Requests"; then
    if echo "$RESPONSE" | grep -q "Retry-After: 1"; then
        echo "✅ rate limiter test passed"
    else
        echo "❌ rate limiter test failed (missing Retry-After)"
        exit 1
    fi
else
    echo "❌ rate limiter test failed (not 429)"
    echo "$RESPONSE"
    exit 1
fi

# Test 6: Circuit Breaker triggering
echo "Triggering circuit breaker (2 failures)..."
curl -s -D - -H "Authorization: Bearer $JWT_TOKEN" -x http://localhost:8080 http://localhost:8081/error > /dev/null
curl -s -D - -H "Authorization: Bearer $JWT_TOKEN" -x http://localhost:8080 http://localhost:8081/error > /dev/null

# Test 7: Circuit Breaker Open
RESPONSE=$(curl -s -D - -H "Authorization: Bearer $JWT_TOKEN" -x http://localhost:8080 http://localhost:8081)
if echo "$RESPONSE" | grep -q "503 Service Unavailable"; then
    echo "✅ Circuit Breaker Open test passed"
else
    echo "❌ Circuit Breaker Open test failed (expected 503)"
    echo "$RESPONSE"
    exit 1
fi

# Test 8: Circuit Breaker Half-Open/Recovery
echo "Waiting 3 seconds for Circuit Breaker timeout..."
sleep 3
RESPONSE=$(curl -s -D - -H "Authorization: Bearer $JWT_TOKEN" -x http://localhost:8080 http://localhost:8081)
if echo "$RESPONSE" | grep -q "200 OK"; then
    echo "✅ Circuit Breaker Half-Open Recovery test passed"
else
    echo "❌ Circuit Breaker Half-Open Recovery test failed (expected 200)"
    echo "$RESPONSE"
    exit 1
fi

echo "==> All E2E tests passed!"
