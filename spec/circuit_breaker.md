# Circuit Breaker

## Overview
The Circuit Breaker is responsible for detecting failures or severe latency in backend servers (upstreams) and temporarily halting requests to them. This prevents cascading failures across the system.

## States
The Circuit Breaker maintains a state machine per upstream server:
- **Closed**: Normal state. Requests are forwarded to the server. If the error threshold is reached, the state transitions to Open.
- **Open**: Requests are blocked. After a specified timeout, the state transitions to Half-Open.
- **Half-Open**: A limited number of requests are allowed to pass through to test if the server has recovered. If successful, it transitions back to Closed. If it fails, it transitions back to Open.

## Configuration
The Circuit Breaker is configured per upstream in `config.yaml`:
```yaml
upstreams:
  - name: my_upstream
    circuit_breaker:
      enabled: true
      error_threshold: 5      # Consecutive errors before opening
      timeout: "30s"          # Duration to stay open before half-open
```

## Behavior
- **Error Tracking**: When proxying a request, if the upstream server returns a 5xx error or a timeout/connection error occurs, the consecutive error count is incremented.
- **Success Tracking**: A successful response (e.g., 200 OK) resets the error count.
- **Health Check Integration**: Servers with an `Open` circuit breaker are treated as Unhealthy and excluded from load balancing routing.

## Implementation Details
1. Add `CircuitBreakerConfig` to `UpstreamConfig`.
2. Implement a `CircuitBreaker` struct with thread-safe state management in `internal/upstream/circuit_breaker.go`.
3. Integrate `CircuitBreaker` into `Server` in `internal/upstream/types.go`.
4. Update `internal/proxy/proxy.go` to record request outcomes in the server's CircuitBreaker.
5. Add E2E tests in `scripts/test-e2e.sh` to verify recovery behavior.
