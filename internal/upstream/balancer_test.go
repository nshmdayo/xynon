package upstream

import (
	"net/http"
	"testing"
)

func TestRoundRobin(t *testing.T) {
	s1, _ := NewServer("http://s1", nil)
	s2, _ := NewServer("http://s2", nil)
	balancer := NewBalancer("round_robin", []*Server{s1, s2})

	req, _ := http.NewRequest("GET", "http://test", nil)

	res1, _ := balancer.NextServer(req)
	if res1.URL.Host != "s1" && res1.URL.Host != "s2" {
		t.Errorf("unexpected host: %s", res1.URL.Host)
	}

	res2, _ := balancer.NextServer(req)
	if res1 == res2 {
		t.Errorf("round robin should return different servers")
	}

	// mark s2 unhealthy
	s2.SetActiveHealthy(false)
	res3, _ := balancer.NextServer(req)
	if res3.URL.Host != "s1" {
		t.Errorf("expected s1 since s2 is unhealthy, got %s", res3.URL.Host)
	}
}

func TestLeastConnections(t *testing.T) {
	s1, _ := NewServer("http://s1", nil)
	s2, _ := NewServer("http://s2", nil)
	balancer := NewBalancer("least_connections", []*Server{s1, s2})

	req, _ := http.NewRequest("GET", "http://test", nil)

	s1.IncConn()
	res1, _ := balancer.NextServer(req)
	if res1.URL.Host != "s2" {
		t.Errorf("expected s2 since it has fewer connections")
	}
}

func TestIPHash(t *testing.T) {
	s1, _ := NewServer("http://s1", nil)
	s2, _ := NewServer("http://s2", nil)
	balancer := NewBalancer("ip_hash", []*Server{s1, s2})

	req, _ := http.NewRequest("GET", "http://test", nil)
	req.RemoteAddr = "192.168.1.1:1234"

	res1, _ := balancer.NextServer(req)
	res2, _ := balancer.NextServer(req)

	if res1 != res2 {
		t.Errorf("ip_hash should return same server for same ip")
	}
}
