# Product Requirements Document: Hot-Reloading

## Feature Description
The hot-reloading mechanism dynamically rebuilds the plugin chain and swaps it automatically when configuration or plugins change without dropping active connections.

## Requirements
- **Monitoring**: Monitor `config.yaml` and the `plugins/` directory via `fsnotify`.
- **Atomic Swapping**: Upon any change, dynamically rebuild the plugin chain and swap it atomically (using `atomic.Pointer`).
