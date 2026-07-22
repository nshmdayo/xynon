with open('docs/requirements/PRD-configuration.md', 'r') as f:
    text = f.read()

text = text.replace(
    '- **Plugin Registry**: Provide a list of plugins defined under `plugins.chain`, specifying the `name`, `type` (`wasm` or `rpc`), and `path`.',
    '- **Plugin Registry**: Provide a list of plugins defined under `plugins.chain`, specifying the `name`, `type` (`wasm` or `rpc`), `path`, and an optional `config` object containing plugin-specific settings.'
)

with open('docs/requirements/PRD-configuration.md', 'w') as f:
    f.write(text)
