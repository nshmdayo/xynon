package admincli

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"github.com/nshmdayo/xynon/internal/config"
	"github.com/nshmdayo/xynon/internal/upstream"
)

func Run(args []string) error {
	if len(args) == 0 {
		return errors.New("admin subcommand requires an action, e.g., 'status'")
	}
	
	switch args[0] {
	case "status":
		return runStatus(args[1:])
	default:
		return fmt.Errorf("unknown admin subcommand: %q", args[0])
	}
}

func runStatus(args []string) error {
	cfg, err := config.Load("config.yaml")
	if err != nil {
		return err
	}
	
	url := fmt.Sprintf("http://%s/_admin/upstreams", cfg.Listen)
	resp, err := http.Get(url)
	if err != nil {
		return fmt.Errorf("failed to query proxy: %w", err)
	}
	defer resp.Body.Close()
	
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("proxy returned status %s", resp.Status)
	}
	
	var statuses []upstream.UpstreamStatus
	if err := json.NewDecoder(resp.Body).Decode(&statuses); err != nil {
		return fmt.Errorf("failed to decode response: %w", err)
	}
	
	for _, u := range statuses {
		fmt.Printf("Upstream: %s\n", u.Name)
		for _, srv := range u.Servers {
			fmt.Printf("  - %s: %s\n", srv.URL, srv.Status)
		}
	}
	
	return nil
}
