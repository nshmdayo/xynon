package plugincli

import (
	"flag"
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/nshmdayo/xynon/internal/config"
)

func runList(args []string, out io.Writer) error {
	fs := flag.NewFlagSet("plugin list", flag.ContinueOnError)
	cfgPath := configFlag(fs)
	if err := mustParse(fs, args, out); err != nil {
		return err
	}

	cfg, err := config.Load(*cfgPath)
	if err != nil {
		return err
	}

	// Build an index of chain-registered plugins by name → order.
	chainIndex := make(map[string]int, len(cfg.Plugins.Chain))
	for i, e := range cfg.Plugins.Chain {
		chainIndex[e.Name] = i + 1
	}

	// Walk plugin directory for .wasm files.
	entries, err := os.ReadDir(cfg.Plugins.Dir)
	if err != nil {
		return fmt.Errorf("cannot read plugin dir %q: %w", cfg.Plugins.Dir, err)
	}

	var found int
	for _, de := range entries {
		if de.IsDir() || filepath.Ext(de.Name()) != ".wasm" {
			continue
		}
		found++
		name := de.Name()[:len(de.Name())-len(".wasm")]
		path := filepath.Join(cfg.Plugins.Dir, de.Name())

		if order, ok := chainIndex[name]; ok {
			fmt.Fprintf(out, "[%d] %s\t%s\n", order, name, path)
		} else {
			fmt.Fprintf(out, "[-] %s\t%s  (not in chain)\n", name, path)
		}
	}

	if found == 0 {
		fmt.Fprintln(out, "no plugins found in", cfg.Plugins.Dir)
	}
	return nil
}
