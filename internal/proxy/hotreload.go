package proxy

import (
	"context"
	"log/slog"
	"os"
	"path/filepath"
	"time"

	"github.com/fsnotify/fsnotify"
	"github.com/nshmdayo/xynon/internal/plugin"
)

const (
	debounce        = 200 * time.Millisecond
	stabilityChecks = 2
	stabilityDelay  = 50 * time.Millisecond
)

// WatcherConfig is the configuration needed to rebuild a chain on file events.
type WatcherConfig struct {
	PluginDir string
	Entries   []plugin.ChainEntry
	Limits    plugin.Limits
	Runtime   *plugin.Runtime
	Registry  *ChainRegistry
}

// Watcher monitors a plugin directory and hot-reloads the chain when WASM
// files change.
type Watcher struct {
	cfg     WatcherConfig
	watcher *fsnotify.Watcher
	done    chan struct{}
}

// NewWatcher creates a Watcher that is ready to call Start().
func NewWatcher(cfg WatcherConfig) (*Watcher, error) {
	fw, err := fsnotify.NewWatcher()
	if err != nil {
		return nil, err
	}
	if err := fw.Add(cfg.PluginDir); err != nil {
		_ = fw.Close()
		return nil, err
	}
	return &Watcher{cfg: cfg, watcher: fw, done: make(chan struct{})}, nil
}

// Start begins watching in a goroutine. It returns immediately.
func (w *Watcher) Start(ctx context.Context) {
	go w.loop(ctx)
}

// Stop closes the watcher goroutine.
func (w *Watcher) Stop() {
	_ = w.watcher.Close()
	<-w.done
}

func (w *Watcher) loop(ctx context.Context) {
	defer close(w.done)

	var timer *time.Timer
	resetTimer := func() {
		if timer != nil {
			timer.Stop()
		}
		timer = time.AfterFunc(debounce, func() {
			w.reload(ctx)
		})
	}

	for {
		select {
		case event, ok := <-w.watcher.Events:
			if !ok {
				return
			}
			if filepath.Ext(event.Name) == ".wasm" {
				slog.Debug("wasm file event", "op", event.Op, "file", event.Name)
				resetTimer()
			}

		case err, ok := <-w.watcher.Errors:
			if !ok {
				return
			}
			slog.Warn("watcher error", "err", err)

		case <-ctx.Done():
			return
		}
	}
}

// reload builds a new chain and swaps it atomically if all plugins load.
func (w *Watcher) reload(ctx context.Context) {
	// Wait for file writes to stabilise.
	entries := w.cfg.Entries
	for i, e := range entries {
		stable, err := waitStable(e.Path)
		if err != nil || !stable {
			slog.Warn("plugin file not stable — aborting reload", "plugin", entries[i].Name, "err", err)
			return
		}
	}

	plugins, err := plugin.LoadChain(ctx, w.cfg.Runtime, entries, w.cfg.Limits)
	if err != nil {
		slog.Warn("hot-reload failed — keeping current chain", "err", err)
		return
	}

	// Close old chain plugins after swap (best effort).
	old := w.cfg.Registry.Load()
	newChain := NewChain(plugins)
	w.cfg.Registry.Store(newChain)
	slog.Info("hot-reload complete", "plugins", len(plugins))

	if old != nil {
		for _, p := range old.Plugins() {
			p.Close(ctx)
		}
	}
}

// waitStable polls the file stats until two consecutive reads agree.
func waitStable(path string) (bool, error) {
	var prev os.FileInfo
	for i := 0; i < stabilityChecks; i++ {
		info, err := os.Stat(path)
		if err != nil {
			if os.IsNotExist(err) {
				// File was deleted — that's a valid, stable state.
				return true, nil
			}
			return false, err
		}
		if prev != nil && info.Size() == prev.Size() && info.ModTime().Equal(prev.ModTime()) {
			return true, nil
		}
		prev = info
		time.Sleep(stabilityDelay)
	}
	return false, nil
}
