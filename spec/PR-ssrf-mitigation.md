# Specification: SSRF Mitigation for Proxy CONNECT Method

## Objective
Prevent Server-Side Request Forgery (SSRF) vulnerabilities in the proxy's `CONNECT` method. Currently, the proxy blindly connects to the host specified in `r.Host` via `net.DialTimeout`. This could allow an external user to proxy traffic into internal/private networks (e.g., `localhost`, `192.168.x.x`, `10.x.x.x`) if the proxy is publicly exposed.

## Scope
This specification is appropriately sized for a single Pull Request. It focuses on adding IP validation before establishing a TCP tunnel.

## Requirements
1. **IP Resolution & Validation:**
   - In `internal/proxy/proxy.go` (`handleTunnel`), before calling `net.DialTimeout`, resolve the hostname in `r.Host` to an IP address.
   - Check if the resolved IP address belongs to a private/loopback network (e.g., using `net.IP.IsPrivate()` and `net.IP.IsLoopback()` in Go 1.17+).
2. **Behavior on Violation:**
   - If the IP is internal/private, the proxy must reject the connection and return an HTTP `403 Forbidden` or `502 Bad Gateway`.
3. **Configuration (Optional but recommended):**
   - Provide a configuration toggle (e.g., `allow_local_network: false` in `Config`) to allow private IPs for testing or internal use cases.
4. **Testing:**
   - Add unit tests in `internal/proxy/proxy_test.go` to verify that `CONNECT localhost:80` is rejected.
   - Add unit tests to verify that normal external addresses (e.g., `example.com:443`) are permitted.
