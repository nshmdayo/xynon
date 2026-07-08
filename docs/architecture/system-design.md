# Xynon System Design

## 1. Overview
Xynon is a high-performance, extensible forward HTTP proxy server written in Go. Its defining feature is a dynamic, hot-reloadable plugin chain that allows users to inspect, mutate, or short-circuit HTTP requests and responses on the fly.

To maximize flexibility and performance, Xynon supports two distinct plugin architectures:
- **WASM Plugins**: Lightweight, in-process plugins executed via `wazero`.
- **RPC Plugins**: Out-of-process plugins executed via `hashicorp/go-plugin`.

## 2. Directory Layout
The workspace is organized as follows:
*   `cmd/xynon/`: Proxy CLI entrypoint (`main.go`).
*   `internal/config/`: Configuration definitions, loader, validator, and serialiser.
*   `internal/plugin/`: Shared `Handler` interface, WASM runtime loader (`wazero`), and RPC client loader (`hashicorp/go-plugin`).
*   `internal/plugin/abi/`: ABI contract definitions (e.g., versioning, host/guest exports/imports).
*   `internal/plugincli/`: Subcommand handlers for plugin CLI operations (`list`, `add`, `remove`, `build`).
*   `internal/proxy/`: Forward HTTP proxy server logic, plugin chain registration, and hot-reload watcher.
*   `examples/plugins/`: Sample Go plugin sources (e.g., `add-header`) compile-ready for TinyGo WASM.
*   `examples/config.yaml`: Default configuration blueprint for running the proxy and setting the plugin chain.
*   `spec/`: Declarative specifications defining required behaviors for features.

## 3. Core Components
### 3.1 The Proxy Engine (`internal/proxy`)
The proxy engine is responsible for receiving incoming HTTP requests, executing the plugin chain, and forwarding requests to the upstream server. 
- **Plugin Chain**: A synchronized list of `plugin.Handler` interfaces. The chain is executed sequentially for each request (`OnRequest`) and in reverse order for each response (`OnResponse`).

### 3.2 Plugin Handlers (`internal/plugin`)
All plugins, regardless of their underlying execution model, implement the `Handler` interface to ensure uniform processing.

### 3.3 Configuration (`internal/config`)
Configurations are defined in a YAML file (e.g., `config.yaml`).

## 4. Development Commands
The project defines helper operations in a `Makefile`. Below are the exact commands to run:

### Build Commands
*   **Build Proxy Host Binary**:
    ```bash
    make build
    ```
    *(Direct command: `CGO_ENABLED=0 go build -o bin/xynon ./cmd/xynon`)*
*   **Build WASM Plugins**:
    ```bash
    make wasm
    ```
    *(Compiles all Go files under `examples/plugins/` using TinyGo targeting WASM)*

### Execution Commands
*   **Start Forward Proxy**:
    ```bash
    ./bin/xynon -config examples/config.yaml
    ```
    *Optionally disable hot reloading:*
    ```bash
    ./bin/xynon -config examples/config.yaml -no-hot-reload
    ```

### Test & Maintenance Commands
*   **Run Test Suite**:
    ```bash
    make test
    ```
    *(Direct command: `go test ./...`)*
*   **Clean Build Artifacts**:
    ```bash
    make clean
    ```

## 5. Coding Standards
AI agents must write code that is clean, readable, and consistent with the existing codebase:

### Naming Conventions
*   **Go packages**: All package names should be lowercase single words (e.g., `config`, `plugin`, `proxy`). Avoid underscores or mixed capitalization.
*   **Go types and functions**: Standard Go `CamelCase` rules. Exported types, fields, and functions must start with an uppercase letter; private ones must be lowercase.

### Coding Best Practices
*   **Gofmt**: All Go files must pass standard `gofmt` (tab indented).
*   **Error Wrapping**: Always use standard error wrapping with `%w` when errors bubble up.
*   **Logging**: Use `log/slog` for structured logging, using camelCase for attributes keys.
*   **Concurrency Safety**: Access to shared mutable states (e.g., the plugin Chain registry or configurations during hot reload) must be synchronized using Go's concurrent primitives (`sync.RWMutex`, `atomic.Pointer`, or channel-based synchronization). Refer to `internal/proxy/chain.go` for implementation references.

## 6. Agent Constraints
When modifying the codebase, the following rules are **strictly mandatory**:
1.  **Preserve Coding Patterns**: Follow existing structure patterns, particularly the separation of CLI packages (`internal/plugincli`) and execution logic. Do not bypass or reimplement logic defined in `internal/plugin/abi`.
2.  **Maintain Documentation Integrity**: Preserve all existing comments and docstrings. Do not remove or alter comments unless requested or directly updating the surrounding code.
3.  **WASM ABI Protection**: Do not change host/guest function signatures or ABI behaviors (`internal/plugin/abi/abi.go`) without explicitly incrementing `CurrentVersion` and ensuring plugins are rebuilt. Any change must maintain compatibility or gracefully reject outdated plugins.
4.  **No CGO**: Keep `CGO_ENABLED=0` to ensure static linking and portability of the generated `xynon` binary.
5.  **Plugin Build Constraints**: WASM plugins must only be built with TinyGo under Go 1.23 environment, avoiding imports of packages that use CGO or dependencies incompatible with WebAssembly targets. RPC plugins can use standard Go, but must implement the expected `hashicorp/go-plugin` interfaces.
6.  **Run Tests Before Finalizing**: Always run `go test ./...` to verify no regression is introduced.
7.  **Auto-Test Hook (AI Behavior)**: Whenever you (the AI agent) implement new features, fix bugs, or modify `.go` files, you MUST automatically execute `make test` via the `run_command` tool before ending your turn. If the tests fail, you must attempt to fix the code and re-run the tests. Do not wait for the user to prompt you to run tests.
