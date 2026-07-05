# Xynon Architecture Specification

## 1. Overview
Xynon is a high-performance, extensible forward HTTP proxy server written in Go. Its defining feature is a dynamic, hot-reloadable plugin chain that allows users to inspect, mutate, or short-circuit HTTP requests and responses on the fly. 

To maximize flexibility and performance, Xynon supports two distinct plugin architectures:
- **WASM Plugins**: Lightweight, in-process plugins executed via `wazero`.
- **RPC Plugins**: Out-of-process plugins executed via `hashicorp/go-plugin`.

## 2. Core Components

### 2.1 The Proxy Engine (`internal/proxy`)
The proxy engine is responsible for receiving incoming HTTP requests, executing the plugin chain, and forwarding requests to the upstream server. 
- **Plugin Chain**: A synchronized list of `plugin.Handler` interfaces. The chain is executed sequentially for each request (`OnRequest`) and in reverse order for each response (`OnResponse`).
- **Hot-Reloading**: The proxy monitors `config.yaml` and the `plugins/` directory via `fsnotify`. Upon any change, it dynamically rebuilds the plugin chain and swaps it atomically (using `atomic.Pointer`) without dropping active connections.

### 2.2 Plugin Handlers (`internal/plugin`)
All plugins, regardless of their underlying execution model, implement the `Handler` interface to ensure uniform processing.

#### WASM Plugins (`WasmHandler`)
- **Runtime**: `wazero` (Zero-dependency WebAssembly runtime).
- **Compilation Target**: `wasip1` (WASI Snapshot Preview 1) via TinyGo.
- **ABI Boundaries**: Communication between the Go host and the WASM guest is done via explicit memory sharing and exported ABI functions (e.g., `xynon_abi_version`, `on_request`, `on_response`). Host functions are injected into the WASM environment to allow guests to manipulate HTTP headers and status codes.

#### RPC Plugins (`HandlerRPCClient` / `HandlerRPCServer`)
- **Runtime**: `hashicorp/go-plugin` over standard `net/rpc`.
- **Execution Model**: Out-of-process. The proxy host launches the RPC plugin as a standalone child process. 
- **Data Passing**: To avoid complex shared-memory management across processes, HTTP headers and statuses are passed by value during the RPC call. The guest mutates the copy and returns the updated state to the host.

### 2.3 Configuration (`internal/config`)
Configurations are defined in a YAML file (e.g., `config.yaml`).
- **Proxy Settings**: `listen` (e.g., `:8080`), `timeout_ms`.
- **Plugin Registry**: A list of plugins defined under `plugins.chain`, specifying the `name`, `type` (`wasm` or `rpc`), and `path`.

### 2.4 CLI Tooling (`internal/plugincli`)
Xynon provides built-in subcommands to manage plugins seamlessly:
- `xynon plugin list`: Displays currently configured plugins.
- `xynon plugin add`: Adds a new plugin to the configuration.
- `xynon plugin remove`: Removes a plugin from the configuration.
- `xynon plugin build`: Compiles a TinyGo WASM plugin from source.

## 3. ABI and Lifecycle Actions
During the execution of `OnRequest` and `OnResponse`, plugins can dictate the proxy's next step by returning an `Action` code:
- `ActionContinue`: Proceed to the next plugin in the chain.
- `ActionStop`: Stop executing the current chain phase, but continue processing the HTTP request/response.

Plugins can also explicitly trigger a **Short-Circuit** during `OnRequest`, immediately halting further plugin execution and proxy forwarding, returning a custom HTTP status code directly to the client.

## 4. Development Constraints
- `CGO_ENABLED=0` must be maintained for the proxy binary.
- WASM plugins must be compiled using Go 1.23+ and TinyGo.
