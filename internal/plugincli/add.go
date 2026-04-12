package plugincli

import (
	"flag"
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/nshmdayo/xynon/internal/config"
)

func runAdd(args []string) error {
	fs := flag.NewFlagSet("plugin add", flag.ContinueOnError)
	cfgPath := configFlag(fs)
	if err := mustParse(fs, args, os.Stderr); err != nil {
		return err
	}
	if fs.NArg() < 1 {
		return fmt.Errorf("usage: xynon plugin add <path-to-wasm> [-config <file>]")
	}

	src := fs.Arg(0)

	cfg, err := config.Load(*cfgPath)
	if err != nil {
		return err
	}

	// Validate source exists before touching anything.
	if _, err := os.Stat(src); err != nil {
		return fmt.Errorf("plugin add: source file %q not found: %w", src, err)
	}

	name := filepath.Base(src)
	name = name[:len(name)-len(filepath.Ext(name))] // strip .wasm
	dst := filepath.Join(cfg.Plugins.Dir, filepath.Base(src))

	// Copy file.
	if err := copyFile(src, dst); err != nil {
		return fmt.Errorf("plugin add: copy: %w", err)
	}

	// Add to chain only if not already registered.
	alreadyIn := false
	for _, e := range cfg.Plugins.Chain {
		if e.Name == name {
			alreadyIn = true
			break
		}
	}
	if !alreadyIn {
		cfg.Plugins.Chain = append(cfg.Plugins.Chain, config.PluginEntry{Name: name})
	}

	if err := config.Save(*cfgPath, cfg); err != nil {
		return err
	}

	if alreadyIn {
		fmt.Printf("updated: %s (already in chain, file overwritten)\n", name)
	} else {
		fmt.Printf("added: %s (chain position %d)\n", name, len(cfg.Plugins.Chain))
	}
	return nil
}

func copyFile(src, dst string) error {
	data, err := os.ReadFile(src)
	if err != nil {
		return err
	}
	return os.WriteFile(dst, data, 0o644)
}

// copyFileWithWriter is used in tests to redirect output.
func runAddWithWriter(args []string, _ io.Writer) error {
	return runAdd(args)
}
