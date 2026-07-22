package upstream

import (
	"errors"
	"net/http"
	"net/url"
	"sync"
	"sync/atomic"
	"time"
)

var (
	ErrNoHealthyBackends = errors.New("no healthy backends available")
)

type Status string

const (
	StatusHealthy   Status = "Healthy"
	StatusUnhealthy Status = "Unhealthy"
)

type ServerStatus struct {
	URL    string `json:"url"`
	Status Status `json:"status"`
}

type UpstreamStatus struct {
	Name    string         `json:"name"`
	Servers []ServerStatus `json:"servers"`
}

type Server struct {
	URL *url.URL

	mu            sync.RWMutex
	activeHealthy bool
	
	// passive checking
	fails            int
	lastPassiveCheck time.Time
	passiveHealthy   bool

	// connection tracking for least_connections
	activeConns atomic.Int64
}

func NewServer(rawURL string) (*Server, error) {
	u, err := url.Parse(rawURL)
	if err != nil {
		return nil, err
	}
	return &Server{
		URL:            u,
		activeHealthy:  true,
		passiveHealthy: true,
	}, nil
}

func (s *Server) IsHealthy() bool {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.activeHealthy && s.passiveHealthy
}

func (s *Server) SetActiveHealthy(healthy bool) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.activeHealthy = healthy
}

func (s *Server) RecordPassiveFailure(maxFails int, failTimeout time.Duration) {
	s.mu.Lock()
	defer s.mu.Unlock()
	now := time.Now()
	if now.Sub(s.lastPassiveCheck) > failTimeout {
		s.fails = 0
	}
	s.fails++
	s.lastPassiveCheck = now
	if s.fails >= maxFails {
		s.passiveHealthy = false
	}
}

func (s *Server) RecordPassiveSuccess(failTimeout time.Duration) {
	s.mu.Lock()
	defer s.mu.Unlock()
	now := time.Now()
	if now.Sub(s.lastPassiveCheck) > failTimeout {
		s.fails = 0
	}
	s.passiveHealthy = true
}

func (s *Server) Status() Status {
	if s.IsHealthy() {
		return StatusHealthy
	}
	return StatusUnhealthy
}

func (s *Server) IncConn() {
	s.activeConns.Add(1)
}

func (s *Server) DecConn() {
	s.activeConns.Add(-1)
}

func (s *Server) ActiveConns() int64 {
	return s.activeConns.Load()
}

type Balancer interface {
	NextServer(req *http.Request) (*Server, error)
	Servers() []*Server
}
