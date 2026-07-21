package lb

import (
	"net/http"
	"time"

	"github.com/nshmdayo/xynon/internal/config"
)

type HealthChecker struct {
	lb     *LoadBalancer
	config config.HealthCheckConfig
	client *http.Client
	stopCh chan struct{}
}

func NewHealthChecker(lb *LoadBalancer, cfg config.HealthCheckConfig) *HealthChecker {
	timeout, _ := time.ParseDuration(cfg.Timeout)
	if timeout == 0 {
		timeout = 2 * time.Second
	}
	return &HealthChecker{
		lb:     lb,
		config: cfg,
		client: &http.Client{Timeout: timeout},
		stopCh: make(chan struct{}),
	}
}

func (hc *HealthChecker) Start() {
	interval, _ := time.ParseDuration(hc.config.Interval)
	if interval == 0 {
		interval = 5 * time.Second
	}
	go func() {
		ticker := time.NewTicker(interval)
		defer ticker.Stop()
		for {
			select {
			case <-ticker.C:
				hc.checkAll()
			case <-hc.stopCh:
				return
			}
		}
	}()
}

func (hc *HealthChecker) Stop() {
	close(hc.stopCh)
}

func (hc *HealthChecker) checkAll() {
	hc.lb.mu.RLock()
	servers := hc.lb.Servers
	hc.lb.mu.RUnlock()

	for _, s := range servers {
		go hc.checkOne(s)
	}
}

func (hc *HealthChecker) checkOne(s *ServerStatus) {
	resp, err := hc.client.Get(s.URL + hc.config.Path)
	alive := false
	if err == nil && resp.StatusCode >= 200 && resp.StatusCode < 400 {
		alive = true
	}
	if resp != nil {
		resp.Body.Close()
	}

	s.Mu.Lock()
	defer s.Mu.Unlock()

	if alive {
		s.Failures = 0
		s.Successes++
		if hc.config.HealthyThreshold <= 0 || s.Successes >= hc.config.HealthyThreshold {
			s.Alive = true
		}
	} else {
		s.Successes = 0
		s.Failures++
		if hc.config.UnhealthyThreshold <= 0 || s.Failures >= hc.config.UnhealthyThreshold {
			s.Alive = false
		}
	}
}
