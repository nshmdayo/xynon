package proxy

import (
	"context"
	"encoding/json"
	"io"
	"log/slog"
	"net"
	"net/http"
	"net/url"
	"sync/atomic"
	"time"

	"github.com/nshmdayo/xynon/internal/config"
	"github.com/nshmdayo/xynon/internal/lb"
	"github.com/nshmdayo/xynon/internal/plugin/abi"
)

// Proxy is an HTTP forward proxy that dispatches to a plugin chain.
type Proxy struct {
	registry          *ChainRegistry
	transport         http.RoundTripper
	allowLocalNetwork bool
	upstreams         map[string]*lb.LoadBalancer
	healthCheckers    []*lb.HealthChecker
}

// New constructs a Proxy using the given chain registry.
func New(registry *ChainRegistry, allowLocalNetwork bool, cfgUpstreams map[string]config.UpstreamConfig) *Proxy {
	upstreams := make(map[string]*lb.LoadBalancer)
	var hcs []*lb.HealthChecker
	for name, uCfg := range cfgUpstreams {
		bal := lb.NewLoadBalancer(uCfg.Algorithm, uCfg.Servers)
		upstreams[name] = bal
		hc := lb.NewHealthChecker(bal, uCfg.HealthCheck)
		hc.Start()
		hcs = append(hcs, hc)
	}

	return &Proxy{
		registry:          registry,
		transport:         http.DefaultTransport,
		allowLocalNetwork: allowLocalNetwork,
		upstreams:         upstreams,
		healthCheckers:    hcs,
	}
}

// Close stops background workers.
func (p *Proxy) Close() {
	for _, hc := range p.healthCheckers {
		hc.Stop()
	}
}

// ServeHTTP handles both plain HTTP proxy requests and HTTPS CONNECT tunnels.
func (p *Proxy) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodConnect {
		p.handleTunnel(w, r)
		return
	}
	p.handleHTTP(w, r)
}

func (p *Proxy) handleAdminUpstreams(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	type AdminServer struct {
		URL         string `json:"url"`
		Alive       bool   `json:"alive"`
		ActiveConns int64  `json:"active_conns"`
	}
	type AdminUpstream struct {
		Algorithm string        `json:"algorithm"`
		Servers   []AdminServer `json:"servers"`
	}

	res := make(map[string]AdminUpstream)
	for name, bal := range p.upstreams {
		var srvs []AdminServer
		servers := bal.Servers
		for _, s := range servers {
			s.Mu.RLock()
			srvs = append(srvs, AdminServer{
				URL:         s.URL,
				Alive:       s.Alive,
				ActiveConns: atomic.LoadInt64(&s.ActiveConns),
			})
			s.Mu.RUnlock()
		}
		res[name] = AdminUpstream{
			Algorithm: bal.Algorithm,
			Servers:   srvs,
		}
	}

	enc := json.NewEncoder(w)
	_ = enc.Encode(res)
}

func (p *Proxy) handleHTTP(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path == "/_admin/upstreams" {
		p.handleAdminUpstreams(w, r)
		return
	}

	// Snapshot the current plugin chain once for this request.
	chain := p.registry.Load()

	// Clone the request because we will mutate its headers and URI,
	// and http.RoundTripper requires the original request to remain unmodified.
	r = r.Clone(r.Context())

	// Run on_request hooks.
	action, shortCircuit, scStatus := p.runOnRequest(r.Context(), chain, r.Header)
	if action == abi.ActionShortCircuit || shortCircuit {
		http.Error(w, http.StatusText(scStatus), scStatus)
		return
	}

	// Remove hop-by-hop headers before forwarding.
	removeHopByHop(r.Header)
	r.RequestURI = ""

	var srv *lb.ServerStatus
	if bal, ok := p.upstreams[r.Host]; ok {
		clientIP, _, _ := net.SplitHostPort(r.RemoteAddr)
		if clientIP == "" {
			clientIP = r.RemoteAddr
		}
		srv = bal.Next(clientIP)
		if srv == nil {
			http.Error(w, "Service Unavailable", http.StatusServiceUnavailable)
			return
		}

		srv.IncConn()
		defer srv.DecConn()

		target, _ := url.Parse(srv.URL)
		if target != nil {
			r.URL.Scheme = target.Scheme
			r.URL.Host = target.Host
			r.Host = target.Host
		}
	}

	resp, err := p.transport.RoundTrip(r)
	if err != nil || (resp != nil && resp.StatusCode >= 500) {
		if srv != nil {
			// For passive health check, we report error
			srv.ReportError(3) // Default unhealthy threshold
		}
	}
	if err != nil {
		http.Error(w, "bad gateway: "+err.Error(), http.StatusBadGateway)
		return
	}
	defer resp.Body.Close()

	// Run on_response hooks.
	p.runOnResponse(r.Context(), chain, resp.Header, resp.StatusCode)

	// Copy response to client.
	removeHopByHop(resp.Header)
	copyHeader(w.Header(), resp.Header)
	w.WriteHeader(resp.StatusCode)
	_, _ = io.Copy(w, resp.Body)
}

func (p *Proxy) handleTunnel(w http.ResponseWriter, r *http.Request) {
	// Snapshot the current plugin chain for this request.
	chain := p.registry.Load()

	// Clone the request because plugins might mutate its headers.
	r = r.Clone(r.Context())

	// Run on_request hooks for CONNECT requests.
	action, shortCircuit, scStatus := p.runOnRequest(r.Context(), chain, r.Header)
	if action == abi.ActionShortCircuit || shortCircuit {
		http.Error(w, http.StatusText(scStatus), scStatus)
		return
	}

	if !p.allowLocalNetwork {
		host, _, err := net.SplitHostPort(r.Host)
		if err != nil {
			host = r.Host
		}

		ips, err := net.DefaultResolver.LookupIP(r.Context(), "ip", host)
		if err == nil {
			for _, ip := range ips {
				if ip.IsPrivate() || ip.IsLoopback() || ip.IsUnspecified() || ip.IsLinkLocalUnicast() || ip.IsLinkLocalMulticast() {
					http.Error(w, "access to private/local network is forbidden", http.StatusForbidden)
					return
				}
			}
		}
	}

	dst, err := net.DialTimeout("tcp", r.Host, 10*time.Second)
	if err != nil {
		http.Error(w, "cannot connect to upstream: "+err.Error(), http.StatusBadGateway)
		return
	}
	defer dst.Close()

	hijacker, ok := w.(http.Hijacker)
	if !ok {
		http.Error(w, "hijacking not supported", http.StatusInternalServerError)
		return
	}
	clientConn, _, err := hijacker.Hijack()
	if err != nil {
		return
	}
	defer clientConn.Close()

	_, _ = clientConn.Write([]byte("HTTP/1.1 200 Connection established\r\n\r\n"))
	go func() { _, _ = io.Copy(dst, clientConn) }()
	_, _ = io.Copy(clientConn, dst)
}

// runOnRequest executes the on_request hook for each plugin in the chain.
// Returns the effective action and short-circuit details.
func (p *Proxy) runOnRequest(ctx context.Context, chain *Chain, header http.Header) (abi.Action, bool, int) {
	for _, pl := range chain.Plugins() {
		action, sc, scStatus, err := pl.OnRequest(ctx, header)
		if err != nil {
			slog.Warn("plugin on_request error — skipping", "plugin", pl.Name(), "err", err)
			continue
		}
		if action == abi.ActionShortCircuit || sc {
			return abi.ActionShortCircuit, true, scStatus
		}
	}
	return abi.ActionContinue, false, 0
}

// runOnResponse executes the on_response hook for each plugin in the chain.
func (p *Proxy) runOnResponse(ctx context.Context, chain *Chain, header http.Header, statusCode int) {
	for _, pl := range chain.Plugins() {
		_, _, _, err := pl.OnResponse(ctx, header, statusCode)
		if err != nil {
			slog.Warn("plugin on_response error — skipping", "plugin", pl.Name(), "err", err)
		}
	}
}

// copyHeader copies src headers into dst.
func copyHeader(dst, src http.Header) {
	for k, vv := range src {
		for _, v := range vv {
			dst.Add(k, v)
		}
	}
}

// hopByHopHeaders is the standard set of hop-by-hop headers per RFC 7230.
var hopByHopHeaders = []string{
	"Connection", "Proxy-Connection", "Keep-Alive", "Proxy-Authenticate",
	"Proxy-Authorization", "Te", "Trailers", "Transfer-Encoding", "Upgrade",
}

// removeHopByHop strips hop-by-hop headers per RFC 7230.
func removeHopByHop(h http.Header) {
	// Respect the Connection header's list of nominated hop-by-hop headers.
	for _, name := range h["Connection"] {
		h.Del(name)
	}
	for _, name := range hopByHopHeaders {
		h.Del(name)
	}
}
