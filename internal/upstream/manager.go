package upstream

import (
	"context"
	"github.com/nshmdayo/xynon/internal/config"
	"sync"
	"time"
)

type Upstream struct {
	Name     string
	Balancer Balancer
	cfg      config.UpstreamConfig
	
	cancel context.CancelFunc
}

func (u *Upstream) Config() config.UpstreamConfig {
	return u.cfg
}

func (u *Upstream) Stop() {
	if u.cancel != nil {
		u.cancel()
	}
}

func (u *Upstream) Status() UpstreamStatus {
	servers := u.Balancer.Servers()
	statuses := make([]ServerStatus, len(servers))
	for i, srv := range servers {
		statuses[i] = ServerStatus{
			URL:    srv.URL.String(),
			Status: srv.Status(),
		}
	}
	return UpstreamStatus{
		Name:    u.Name,
		Servers: statuses,
	}
}

type Manager struct {
	mu        sync.RWMutex
	upstreams map[string]*Upstream
}

func NewManager(ctx context.Context, cfg []config.UpstreamConfig) (*Manager, error) {
	m := &Manager{
		upstreams: make(map[string]*Upstream),
	}
	for _, ucfg := range cfg {
		var servers []*Server

		for _, scfg := range ucfg.Servers {
			var serverCB *CircuitBreaker
			if ucfg.CircuitBreaker.Enabled {
				timeout, err := time.ParseDuration(ucfg.CircuitBreaker.Timeout)
				if err != nil || timeout <= 0 {
					timeout = 30 * time.Second
				}
				threshold := ucfg.CircuitBreaker.ErrorThreshold
				if threshold <= 0 {
					threshold = 5
				}
				serverCB = NewCircuitBreaker(threshold, timeout)
			}

			srv, err := NewServer(scfg.URL, serverCB)
			if err != nil {
				return nil, err
			}
			servers = append(servers, srv)
		}
		
		balancer := NewBalancer(ucfg.Algorithm, servers)
		
		ctx, cancel := context.WithCancel(ctx)
		
		up := &Upstream{
			Name:     ucfg.Name,
			Balancer: balancer,
			cfg:      ucfg,
			cancel:   cancel,
		}
		
		if ucfg.HealthCheck.Active.Enabled {
			for _, srv := range servers {
				hc := NewActiveHealthChecker(srv, ucfg.HealthCheck.Active)
				go hc.Start(ctx)
			}
		}
		
		m.upstreams[ucfg.Name] = up
	}
	
	return m, nil
}

func (m *Manager) Get(name string) *Upstream {
	return m.upstreams[name]
}

func (m *Manager) StopAll() {
	m.mu.Lock()
	defer m.mu.Unlock()
	for _, up := range m.upstreams {
		up.Stop()
	}
}

func (m *Manager) Statuses() []UpstreamStatus {
	m.mu.RLock()
	defer m.mu.RUnlock()
	var res []UpstreamStatus
	for _, up := range m.upstreams {
		res = append(res, up.Status())
	}
	return res
}
