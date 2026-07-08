# Specification: Enable Plugin Hooks for CONNECT Requests

## Objective
Currently, the proxy engine routes standard HTTP requests through the plugin chain (WASM/RPC plugins), allowing plugins to intercept and authorize traffic. However, `CONNECT` requests (HTTPS tunnels) bypass the plugin chain entirely. This specification extends the plugin system to intercept `CONNECT` requests before the TCP tunnel is established.

## Scope
This specification is appropriately sized for a single Pull Request. It focuses on executing the `OnRequest` plugin hooks during the `CONNECT` phase. (Note: `OnResponse` hooks do not apply to opaque TCP tunnels).

## Requirements
1. **Execute Plugin Chain:**
   - In `internal/proxy/proxy.go` (`handleTunnel`), retrieve the active plugin chain using `p.registry.Load()`.
   - Execute `p.runOnRequest(r.Context(), chain, r.Header)` before attempting to dial the upstream host.
2. **Short-Circuit Handling:**
   - If any plugin returns an `ActionShortCircuit` (e.g., the `auth` plugin denies access), the proxy must abort the tunnel creation.
   - The proxy should return the HTTP status code specified by the plugin (e.g., `401 Unauthorized` or `403 Forbidden`).
3. **Testing:**
   - Extend the E2E test script (`scripts/test-e2e.sh`) to perform a `CONNECT` request against the proxy.
   - Verify that an unauthenticated `CONNECT` request is correctly rejected by the `auth` plugin.
   - Verify that an authenticated `CONNECT` request successfully establishes a tunnel.
