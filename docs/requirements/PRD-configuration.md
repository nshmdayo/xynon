# Product Requirements Document: Configuration

## Feature Description
Configurations are defined in a YAML file (e.g., `config.yaml`) that specifies the proxy's runtime behavior and the plugin registry.

## Requirements
- **Proxy Settings**: Define proxy-level settings such as `listen` (e.g., `:8080`) and `timeout_ms`.
- **Plugin Registry**: Provide a list of plugins defined under `plugins.chain`, specifying the `name`, `type` (`wasm` or `rpc`), and `path`.
- **Tags**: Struct fields in configuration files should have snake_case tags, e.g., `yaml:"timeout_ms"`.
