package lb

import (
	"crypto/sha256"
	"encoding/binary"
	"sync"
	"sync/atomic"
)

type ServerStatus struct {
	URL          string
	Alive        bool
	ActiveConns  int64
	Mu           sync.RWMutex
	Failures     int
	Successes    int
}

func (s *ServerStatus) IsAlive() bool {
	s.Mu.RLock()
	defer s.Mu.RUnlock()
	return s.Alive
}

type LoadBalancer struct {
	Algorithm string
	Servers   []*ServerStatus
	mu        sync.RWMutex
	idx       uint64
}

func NewLoadBalancer(algorithm string, urls []string) *LoadBalancer {
	var servers []*ServerStatus
	for _, u := range urls {
		servers = append(servers, &ServerStatus{
			URL:   u,
			Alive: true,
		})
	}
	return &LoadBalancer{
		Algorithm: algorithm,
		Servers:   servers,
	}
}

func (lb *LoadBalancer) Next(clientIP string) *ServerStatus {
	lb.mu.RLock()
	defer lb.mu.RUnlock()

	var alive []*ServerStatus
	for _, s := range lb.Servers {
		s.Mu.RLock()
		if s.Alive {
			alive = append(alive, s)
		}
		s.Mu.RUnlock()
	}

	if len(alive) == 0 {
		return nil
	}

	switch lb.Algorithm {
	case "least_connections":
		var best *ServerStatus
		var min int64 = -1
		for _, s := range alive {
			c := atomic.LoadInt64(&s.ActiveConns)
			if min == -1 || c < min {
				min = c
				best = s
			}
		}
		return best
	case "ip_hash":
		h := sha256.Sum256([]byte(clientIP))
		val := binary.BigEndian.Uint64(h[:8])
		return alive[val%uint64(len(alive))]
	case "round_robin":
		fallthrough
	default:
		i := atomic.AddUint64(&lb.idx, 1)
		return alive[i%uint64(len(alive))]
	}
}

func (s *ServerStatus) IncConn() {
	atomic.AddInt64(&s.ActiveConns, 1)
}

func (s *ServerStatus) DecConn() {
	atomic.AddInt64(&s.ActiveConns, -1)
}

func (s *ServerStatus) ReportError(unhealthyThreshold int) {
	s.Mu.Lock()
	defer s.Mu.Unlock()
	s.Failures++
	s.Successes = 0
	if s.Failures >= unhealthyThreshold {
		s.Alive = false
	}
}
