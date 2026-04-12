package plugincli

import (
	"flag"
	"fmt"
	"os"

	"github.com/nshmdayo/xynon/internal/config"
)

func runRemove(args []string) error {
	fs := flag.NewFlagSet("plugin remove", flag.ContinueOnError)
	cfgPath := configFlag(fs)
	if err := mustParse(fs, args, os.Stderr); err != nil {
		return err
	}
	if fs.NArg() < 1 {
		return fmt.Errorf("usage: xynon plugin remove <name> [-config <file>]")
	}
	name := fs.Arg(0)

	cfg, err := config.Load(*cfgPath)
	if err != nil {
		return err
	}

	newChain := cfg.Plugins.Chain[:0:0]
	found := false
	for _, e := range cfg.Plugins.Chain {
		if e.Name == name {
			found = true
			continue
		}
		newChain = append(newChain, e)
	}
	if !found {
		return fmt.Errorf("plugin remove: %q is not in the chain", name)
	}

	cfg.Plugins.Chain = newChain
	if err := config.Save(*cfgPath, cfg); err != nil {
		return err
	}

	fmt.Printf("removed: %s (wasm file not deleted)\n", name)
	return nil
}
