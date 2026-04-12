// Package plugincli implements the `xynon plugin` subcommand group.
package plugincli

import (
	"flag"
	"fmt"
	"io"
	"os"
)

// Run dispatches the plugin subcommand from args (os.Args[2:] equivalent).
// It writes output to stdout and errors to stderr.
func Run(args []string) error {
	if len(args) == 0 {
		return usageError()
	}

	sub := args[0]
	rest := args[1:]

	switch sub {
	case "list":
		return runList(rest, os.Stdout)
	case "add":
		return runAdd(rest)
	case "remove":
		return runRemove(rest)
	case "build":
		return runBuild(rest)
	default:
		return fmt.Errorf("unknown plugin subcommand %q\n%s", sub, pluginUsage)
	}
}

const pluginUsage = `Usage: xynon plugin <subcommand> [flags]

Subcommands:
  list    List plugins and their chain status
  add     Copy a WASM file into the plugin directory and register it
  remove  Remove a plugin from the chain (file is not deleted)
  build   Build a TinyGo plugin source directory into a WASM file
`

func usageError() error {
	return fmt.Errorf("plugin subcommand required\n%s", pluginUsage)
}

// configFlag adds a -config flag to fs and returns a pointer to its value.
func configFlag(fs *flag.FlagSet) *string {
	return fs.String("config", "config.yaml", "path to config file")
}

// mustParse parses fs with args and returns any non-flag error. It prints
// usage to w on -help.
func mustParse(fs *flag.FlagSet, args []string, w io.Writer) error {
	fs.SetOutput(w)
	return fs.Parse(args)
}
