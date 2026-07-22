package main

import (
	"context"
	"flag"
	"log/slog"
	"net/http"

	"github.com/prometheus/client_golang/prometheus/promhttp"
	"os"
	"encoding/json"
	"os/signal"
	"path/filepath"
	"syscall"

	"github.com/nshmdayo/xynon/internal/admincli"
	"github.com/nshmdayo/xynon/internal/config"
	"github.com/nshmdayo/xynon/internal/plugin"
	"github.com/nshmdayo/xynon/internal/plugincli"
	"github.com/nshmdayo/xynon/internal/proxy"
	"github.com/nshmdayo/xynon/internal/upstream"
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

	// Dispatch `xynon admin <subcommand>` before parsing proxy flags.
	if len(os.Args) >= 2 && os.Args[1] == "admin" {
		if err := admincli.Run(os.Args[2:]); err != nil {
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
		path := e.Path
		if path == "" {
			path = filepath.Join(cfg.Plugins.Dir, e.Name)
			if typ == config.PluginTypeWasm {
				path += ".wasm"
			}
		}
		var cfgBytes []byte
		if e.Config != nil {
			cfgBytes, _ = json.Marshal(e.Config)
		}
		entries = append(entries, plugin.ChainEntry{
			Name:   e.Name,
			Type:   typ,
			Path:   path,
			Config: cfgBytes,
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
			PluginDir:  cfg.Plugins.Dir,
			ConfigPath: *configPath,
			ReloadConfig: func() ([]plugin.ChainEntry, error) {
				newCfg, err := config.Load(*configPath)
				if err != nil {
					return nil, err
				}
				newEntries := make([]plugin.ChainEntry, 0, len(newCfg.Plugins.Chain))
				for _, e := range newCfg.Plugins.Chain {
					typ := e.Type
					if typ == "" {
						typ = config.PluginTypeWasm
					}
					path := e.Path
					if path == "" {
						path = filepath.Join(newCfg.Plugins.Dir, e.Name)
						if typ == config.PluginTypeWasm {
							path += ".wasm"
						}
					}
					var cfgBytes []byte
					if e.Config != nil {
						cfgBytes, _ = json.Marshal(e.Config)
					}
					newEntries = append(newEntries, plugin.ChainEntry{
						Name:   e.Name,
						Type:   typ,
						Path:   path,
						Config: cfgBytes,
					})
				}
				return newEntries, nil
			},
			Entries:  entries,
			Limits:   limits,
			Runtime:  rt,
			Registry: reg,
		}
		watcher, err = proxy.NewWatcher(wcfg)
		if err != nil {
			slog.Warn("hot-reload watcher failed to start — running without hot-reload", "err", err)
		} else {
			watcher.Start(ctx)
			slog.Info("hot-reload enabled", "dir", cfg.Plugins.Dir)
		}
	}

	// Start metrics server if enabled
	if cfg.Metrics.Enabled {
		addr := cfg.Metrics.Address
		if addr == "" {
			addr = ":9090"
		}
		go func() {
			metricsMux := http.NewServeMux()
			metricsMux.Handle("/metrics", promhttp.Handler())
			slog.Info("metrics server listening", "addr", addr)
			if err := http.ListenAndServe(addr, metricsMux); err != nil && err != http.ErrServerClosed {
				slog.Error("metrics server failed", "err", err)
			}
		}()
	}

	// Initialize upstream manager.
	upManager, err := upstream.NewManager(ctx, cfg.Upstreams)
	if err != nil {
		slog.Error("failed to init upstreams", "err", err)
		os.Exit(1)
	}

	// Start proxy server.
	p := proxy.New(reg, cfg.AllowLocalNetwork, upManager)
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
		if upManager != nil {
			upManager.StopAll()
		}
		_ = srv.Shutdown(context.Background())
	}()

	slog.Info("proxy listening", "addr", cfg.Listen)
	if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		slog.Error("server error", "err", err)
		os.Exit(1)
	}
}
