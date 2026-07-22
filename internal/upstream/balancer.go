package upstream

import (
	"hash/fnv"
	"net"
	"net/http"
	"sync/atomic"
)

type baseBalancer struct {
	servers []*Server
}

func (b *baseBalancer) Servers() []*Server {
	return b.servers
}

type RoundRobin struct {
	baseBalancer
	next atomic.Uint32
}

func NewRoundRobin(servers []*Server) *RoundRobin {
	return &RoundRobin{baseBalancer: baseBalancer{servers: servers}}
}

func (rr *RoundRobin) NextServer(_ *http.Request) (*Server, error) {
	if len(rr.servers) == 0 {
		return nil, ErrNoHealthyBackends
	}
	
	// try all servers once
	for i := 0; i < len(rr.servers); i++ {
		idx := rr.next.Add(1) % uint32(len(rr.servers))
		srv := rr.servers[idx]
		if srv.IsHealthy() {
			return srv, nil
		}
	}
	
	return nil, ErrNoHealthyBackends
}

type LeastConnections struct {
	baseBalancer
}

func NewLeastConnections(servers []*Server) *LeastConnections {
	return &LeastConnections{baseBalancer: baseBalancer{servers: servers}}
}

func (lc *LeastConnections) NextServer(_ *http.Request) (*Server, error) {
	if len(lc.servers) == 0 {
		return nil, ErrNoHealthyBackends
	}
	
	var best *Server
	var minConns int64 = -1
	
	for _, srv := range lc.servers {
		if !srv.IsHealthy() {
			continue
		}
		conns := srv.ActiveConns()
		if best == nil || conns < minConns {
			best = srv
			minConns = conns
		}
	}
	
	if best == nil {
		return nil, ErrNoHealthyBackends
	}
	return best, nil
}

type IPHash struct {
	baseBalancer
}

func NewIPHash(servers []*Server) *IPHash {
	return &IPHash{baseBalancer: baseBalancer{servers: servers}}
}

func (ih *IPHash) NextServer(req *http.Request) (*Server, error) {
	if len(ih.servers) == 0 {
		return nil, ErrNoHealthyBackends
	}
	
	healthyCount := 0
	for _, srv := range ih.servers {
		if srv.IsHealthy() {
			healthyCount++
		}
	}
	
	if healthyCount == 0 {
		return nil, ErrNoHealthyBackends
	}
	
	ip := req.Header.Get("X-Forwarded-For")
	if ip == "" {
		ip = req.Header.Get("X-Real-IP")
	}
	if ip == "" {
		ip, _, _ = net.SplitHostPort(req.RemoteAddr)
		if ip == "" {
			ip = req.RemoteAddr
		}
	}
	
	h := fnv.New32a()
	h.Write([]byte(ip))
	idx := h.Sum32() % uint32(healthyCount)
	
	for _, srv := range ih.servers {
		if srv.IsHealthy() {
			if idx == 0 {
				return srv, nil
			}
			idx--
		}
	}
	
	return nil, ErrNoHealthyBackends
}

func NewBalancer(algo string, servers []*Server) Balancer {
	switch algo {
	case "least_connections":
		return NewLeastConnections(servers)
	case "ip_hash":
		return NewIPHash(servers)
	case "round_robin":
		fallthrough
	default:
		return NewRoundRobin(servers)
	}
}
