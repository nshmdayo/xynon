# Product Requirements Document: Configuration

## Feature Description
Configurations are defined in a YAML file (e.g., `config.yaml`) that specifies the proxy's runtime behavior and the plugin registry.

## Requirements
- **Proxy Settings**: Define proxy-level settings such as `listen` (e.g., `:8080`) and `timeout_ms`.
- **Plugin Registry**: Provide a list of plugins defined under `plugins.chain`, specifying the `name`, `type` (`wasm` or `rpc`), `path`, and an optional `config` object containing plugin-specific settings.
- **Tags**: Struct fields in configuration files should have snake_case tags, e.g., `yaml:"timeout_ms"`.
