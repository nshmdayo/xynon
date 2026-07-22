# API and Contract Specification

## 1. ABI Boundaries
Communication between the Go host and the WASM guest is done via explicit memory sharing and exported ABI functions. Host functions are injected into the WASM environment to allow guests to manipulate HTTP headers and status codes.

### Exported Functions
Exported functions meant to be matched by the guest/plugin ABI must use snake_case matching the ABI expectations:
- `xynon_abi_version`
- `get_plugin_config` (Host function)
- `on_request`
- `on_response`

## 2. ABI and Lifecycle Actions
During the execution of `OnRequest` and `OnResponse`, plugins can dictate the proxy's next step by returning an `Action` code:
- `ActionContinue`: Proceed to the next plugin in the chain.
- `ActionShortCircuit`: Stop the chain; host uses the short-circuit response (halting further plugin execution and proxy forwarding).
- `ActionError`: Plugin encountered an error; host skips this plugin for this request.

## 3. Configuration Contract
Struct fields in configuration files should have snake_case tags, e.g., `yaml:"timeout_ms"`.
