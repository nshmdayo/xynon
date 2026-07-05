# Agent Rules and Workspace Guidelines

This document outlines the project profile, development workflows, coding standards, and agent constraints for the **Xynon** repository. Any AI coding assistant (like Antigravity) must strictly adhere to these rules when working in this workspace.

---

## 1. Project Profile

### Technology Stack
*   **Primary Language**: Go (v1.26.1+)
*   **WASM Plugin Language**: Go compiled via **TinyGo** (v0.40.1 or compatible, using Go 1.23 for compilation compatibility)
*   **RPC Plugin Language**: Any language supporting `hashicorp/go-plugin` (typically standard Go 1.23+), running out-of-process.
*   **Compilation Target**: WebAssembly (`wasm` / `wasip1`) for WASM plugins, native executables for RPC plugins.

### Architecture & Key Libraries
*   **WebAssembly Runtime**: `github.com/tetratelabs/wazero v1.11.0` (Zero-dependency WebAssembly runtime for Go).
*   **RPC Plugin Runtime**: `github.com/hashicorp/go-plugin v1.8.0` (Manages out-of-process plugin lifecycles and communications).
*   **File System Monitoring**: `github.com/fsnotify/fsnotify v1.9.0` (Used for hot-reloading WebAssembly plugins upon file changes).
*   **Configuration File Parser**: `gopkg.in/yaml.v3 v3.0.1` (Handles loading/saving `config.yaml`).
*   **Structured Logging**: Standard Go `log/slog` library.

---

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
*   `openspec/`: Declarative specifications defining required behaviors for features.

---

## 3. Development Commands

The project defines helper operations in a `Makefile`. Below are the exact commands to run:

### Build Commands
*   **Build Proxy Host Binary**:
    ```bash
    make build
    ```
    *(Direct command: `CGO_ENABLED=0 go build -o bin/xynon ./cmd/xynon`)*
*   **Build WASM Plugins**:
    ```bash
    make plugins
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
*   **Plugin CLI Management**:
    *   **List Plugins**: `./bin/xynon plugin list -config examples/config.yaml`
    *   **Add Plugin**: `./bin/xynon plugin add -config examples/config.yaml ./path/to/plugin.wasm`
    *   **Remove Plugin**: `./bin/xynon plugin remove -config examples/config.yaml plugin-name`
    *   **Build Plugin**: `./bin/xynon plugin build -config examples/config.yaml ./examples/plugins/plugin-name`

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

---

## 4. Coding Standards

AI agents must write code that is clean, readable, and consistent with the existing codebase:

### Naming Conventions
*   **Go packages**: All package names should be lowercase single words (e.g., `config`, `plugin`, `proxy`). Avoid underscores or mixed capitalization.
*   **Go types and functions**: Standard Go `CamelCase` rules. Exported types, fields, and functions must start with an uppercase letter; private ones must be lowercase.
*   **YAML tags**: Struct fields in configuration files should have snake_case tags, e.g., `yaml:"timeout_ms"`.
*   **WASM exports**: Exported functions meant to be matched by the guest/plugin ABI must use snake_case matching the ABI expectations (e.g., `xynon_abi_version`, `on_request`).

### Coding Best Practices
*   **Gofmt**: All Go files must pass standard `gofmt` (tab indented).
*   **Error Wrapping**: Always use standard error wrapping with `%w` when errors bubble up (e.g., `fmt.Errorf("config: cannot read %q: %w", path, err)`).
*   **Logging**: Use `log/slog` for structured logging, using camelCase for attributes keys.
*   **Concurrency Safety**: Access to shared mutable states (e.g., the plugin Chain registry or configurations during hot reload) must be synchronized using Go's concurrent primitives (`sync.RWMutex`, `atomic.Pointer`, or channel-based synchronization). Refer to `internal/proxy/chain.go` for implementation references.

---

## 5. Agent Constraints

When modifying the codebase, the following rules are **strictly mandatory**:

1.  **Preserve Coding Patterns**: Follow existing structure patterns, particularly the separation of CLI packages (`internal/plugincli`) and execution logic. Do not bypass or reimplement logic defined in `internal/plugin/abi`.
2.  **Maintain Documentation Integrity**: Preserve all existing comments and docstrings. Do not remove or alter comments unless requested or directly updating the surrounding code.
3.  **WASM ABI Protection**: Do not change host/guest function signatures or ABI behaviors (`internal/plugin/abi/abi.go`) without explicitly incrementing `CurrentVersion` and ensuring plugins are rebuilt. Any change must maintain compatibility or gracefully reject outdated plugins.
4.  **No CGO**: Keep `CGO_ENABLED=0` to ensure static linking and portability of the generated `xynon` binary.
5.  **Plugin Build Constraints**: WASM plugins must only be built with TinyGo under Go 1.23 environment, avoiding imports of packages that use CGO or dependencies incompatible with WebAssembly targets. RPC plugins can use standard Go, but must implement the expected `hashicorp/go-plugin` interfaces.
6.  **Run Tests Before Finalizing**: Always run `go test ./...` to verify no regression is introduced.
