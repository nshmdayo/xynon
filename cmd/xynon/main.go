package main

import (
	"context"
	"flag"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"

	"github.com/nshmdayo/xynon/internal/config"
	"github.com/nshmdayo/xynon/internal/plugin"
	"github.com/nshmdayo/xynon/internal/plugincli"
	"github.com/nshmdayo/xynon/internal/proxy"
)

func main() {
	// Dispatch `xynon plugin <subcommand>` before parsing proxy flags.
	if len(os.Args) >= 2 && os.Args[1] == "plugin" {
		if err := plugincli.Run(os.Args[2:]); err != nil {
			slog.Error(err.Error())
			os.Exit(1)
		}
		return
	}

	configPath := flag.String("config", "config.yaml", "path to config file")
	noHotReload := flag.Bool("no-hot-reload", false, "disable hot-reload and keep initial chain fixed")
	flag.Parse()

	cfg, err := config.Load(*configPath)
	if err != nil {
		slog.Error("failed to load config", "err", err)
		os.Exit(1)
	}

	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer cancel()

	// Initialise the WASM runtime.
	rt, err := plugin.NewRuntime(ctx)
	if err != nil {
		slog.Error("failed to init plugin runtime", "err", err)
		os.Exit(1)
	}
	defer rt.Close(context.Background())

	// Build chain entries from config.
	entries := make([]plugin.ChainEntry, 0, len(cfg.Plugins.Chain))
	for _, e := range cfg.Plugins.Chain {
		typ := e.Type
		if typ == "" {
			typ = config.PluginTypeWasm
		}
		path := filepath.Join(cfg.Plugins.Dir, e.Name)
		if typ == config.PluginTypeWasm {
			path += ".wasm"
		}
		entries = append(entries, plugin.ChainEntry{
			Name: e.Name,
			Type: typ,
			Path: path,
		})
	}

	limits := plugin.DefaultLimits()

	// Load initial chain.
	plugins, err := plugin.LoadChain(ctx, rt, entries, limits)
	if err != nil {
		slog.Error("failed to load initial plugin chain", "err", err)
		os.Exit(1)
	}

	reg := &proxy.ChainRegistry{}
	reg.Store(proxy.NewChain(plugins))
	slog.Info("initial chain loaded", "plugins", len(plugins))

	// Start hot-reload watcher unless disabled.
	var watcher *proxy.Watcher
	if !*noHotReload {
		wcfg := proxy.WatcherConfig{
			PluginDir: cfg.Plugins.Dir,
			Entries:   entries,
			Limits:    limits,
			Runtime:   rt,
			Registry:  reg,
		}
		watcher, err = proxy.NewWatcher(wcfg)
		if err != nil {
			slog.Warn("hot-reload watcher failed to start — running without hot-reload", "err", err)
		} else {
			watcher.Start(ctx)
			slog.Info("hot-reload enabled", "dir", cfg.Plugins.Dir)
		}
	}

	// Start proxy server.
	p := proxy.New(reg)
	srv := &http.Server{
		Addr:    cfg.Listen,
		Handler: p,
	}

	go func() {
		<-ctx.Done()
		slog.Info("shutting down")
		if watcher != nil {
			watcher.Stop()
		}
		_ = srv.Shutdown(context.Background())
	}()

	slog.Info("proxy listening", "addr", cfg.Listen)
	if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		slog.Error("server error", "err", err)
		os.Exit(1)
	}
}
