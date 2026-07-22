package upstream

import (
	"context"
	"net/http"
	"time"

	"github.com/nshmdayo/xynon/internal/config"
)

type ActiveHealthChecker struct {
	server *Server
	cfg    config.ActiveCheckConfig
	client *http.Client
}

func NewActiveHealthChecker(server *Server, cfg config.ActiveCheckConfig) *ActiveHealthChecker {
	timeout, err := time.ParseDuration(cfg.Timeout)
	if err != nil || timeout == 0 {
		timeout = 2 * time.Second
	}
	return &ActiveHealthChecker{
		server: server,
		cfg:    cfg,
		client: &http.Client{
			Timeout: timeout,
		},
	}
}

func (hc *ActiveHealthChecker) Start(ctx context.Context) {
	interval, err := time.ParseDuration(hc.cfg.Interval)
	if err != nil || interval == 0 {
		interval = 10 * time.Second
	}
	
	ticker := time.NewTicker(interval)
	defer ticker.Stop()
	
	fails := 0
	
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			url := hc.server.URL.String() + hc.cfg.Path
			req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
			if err != nil {
				continue
			}
			
			resp, err := hc.client.Do(req)
			if err != nil || resp.StatusCode != hc.cfg.ExpectedStatus {
				fails++
				if fails >= hc.cfg.MaxFails {
					hc.server.SetActiveHealthy(false)
				}
				if resp != nil {
					resp.Body.Close()
				}
			} else {
				fails = 0
				hc.server.SetActiveHealthy(true)
				resp.Body.Close()
			}
		}
	}
}
