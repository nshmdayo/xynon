package upstreamscli

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"text/tabwriter"
)

type AdminServer struct {
	URL         string `json:"url"`
	Alive       bool   `json:"alive"`
	ActiveConns int64  `json:"active_conns"`
}
type AdminUpstream struct {
	Algorithm string        `json:"algorithm"`
	Servers   []AdminServer `json:"servers"`
}

func Run(args []string) error {
	if len(args) == 0 || args[0] != "status" {
		return fmt.Errorf("usage: xynon upstreams status")
	}

	// Assuming default localhost:8080 or read from config?
	// The prompt didn't specify how to know the admin endpoint address,
	// standard practice for CLI is flag or default. Let's use a simple default.
	adminURL := "http://localhost:8080/_admin/upstreams"

	resp, err := http.Get(adminURL)
	if err != nil {
		return fmt.Errorf("failed to fetch status: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("unexpected status: %s", resp.Status)
	}

	var res map[string]AdminUpstream
	if err := json.NewDecoder(resp.Body).Decode(&res); err != nil {
		return fmt.Errorf("failed to decode response: %w", err)
	}

	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	fmt.Fprintln(w, "UPSTREAM\tALGORITHM\tSERVER\tSTATUS\tCONNECTIONS")
	for name, upstream := range res {
		for _, s := range upstream.Servers {
			status := "DOWN"
			if s.Alive {
				status = "UP"
			}
			fmt.Fprintf(w, "%s\t%s\t%s\t%s\t%d\n", name, upstream.Algorithm, s.URL, status, s.ActiveConns)
		}
	}
	w.Flush()
	return nil
}
