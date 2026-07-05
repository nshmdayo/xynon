package plugincli

import (
	"flag"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/nshmdayo/xynon/internal/config"
)

func runBuild(args []string) error {
	fs := flag.NewFlagSet("plugin build", flag.ContinueOnError)
	cfgPath := configFlag(fs)
	if err := mustParse(fs, args, os.Stderr); err != nil {
		return err
	}
	if fs.NArg() < 1 {
		return fmt.Errorf("usage: xynon plugin build <src-dir> [-config <file>]")
	}
	srcDir := fs.Arg(0)

	cfg, err := config.Load(*cfgPath)
	if err != nil {
		return err
	}

	tinygo, err := resolveTinyGo()
	if err != nil {
		return err
	}

	name := filepath.Base(srcDir)
	outPath := filepath.Join(cfg.Plugins.Dir, name+".wasm")

	cmd := exec.Command(tinygo, "build", "-target", "wasip1", "-o", outPath, srcDir)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	// Allow overriding GOROOT for TinyGo (needed when Go toolchain > 1.25).
	if goroot := os.Getenv("XYNON_TINYGO_GOROOT"); goroot != "" {
		cmd.Env = append(os.Environ(), "GOROOT="+goroot)
	}

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("plugin build: tinygo exited with error: %w", err)
	}

	fmt.Printf("built: %s → %s\n", name, outPath)
	return nil
}

// resolveTinyGo returns the path to the tinygo binary, checking the TINYGO
// environment variable first, then PATH.
func resolveTinyGo() (string, error) {
	if p := os.Getenv("TINYGO"); p != "" {
		if _, err := os.Stat(p); err == nil {
			return p, nil
		}
		return "", fmt.Errorf("plugin build: TINYGO=%q not found", p)
	}

	p, err := exec.LookPath("tinygo")
	if err != nil {
		return "", fmt.Errorf(
			"plugin build: tinygo not found in PATH\n" +
				"Install TinyGo from https://tinygo.org/getting-started/install/ " +
				"or set the TINYGO environment variable to its binary path",
		)
	}
	return p, nil
}
