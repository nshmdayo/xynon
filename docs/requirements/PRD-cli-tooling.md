# Product Requirements Document: CLI Tooling

## Feature Description
Xynon provides built-in subcommands to manage plugins seamlessly via the plugin CLI interface.

## Requirements
- **List Plugins**: `xynon plugin list` - Displays currently configured plugins. Example: `./bin/xynon plugin list -config examples/config.yaml`
- **Add Plugin**: `xynon plugin add` - Adds a new plugin to the configuration. Example: `./bin/xynon plugin add -config examples/config.yaml ./path/to/plugin.wasm`
- **Remove Plugin**: `xynon plugin remove` - Removes a plugin from the configuration. Example: `./bin/xynon plugin remove -config examples/config.yaml plugin-name`
- **Build Plugin**: `xynon plugin build` - Compiles a TinyGo WASM plugin from source. Example: `./bin/xynon plugin build -config examples/config.yaml ./examples/plugins/plugin-name`
