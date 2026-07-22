# Xynon

An HTTP forward proxy written in Go. You can extend request/response processing with **TinyGo WASM plugins** or **RPC plugins**. The proxy supports **hot-reloading**, allowing you to replace plugins without restarting the proxy.

## Architecture

```text
cmd/xynon/           <- Proxy CLI entrypoint
internal/config/     <- Configuration loader
internal/plugin/     <- WASM runtime, RPC loader, ABI
internal/plugin/abi/ <- ABI constants & definitions
internal/proxy/      <- Proxy handler, chain registry, hot-reloading
internal/upstream/   <- Routing, health checks, circuit breakers
examples/plugins/    <- Sample plugins
examples/config.yaml <- Sample configuration
```

## Build

### Core Proxy

```bash
CGO_ENABLED=0 go build -o bin/xynon ./cmd/xynon
```

### WASM Plugins (requires TinyGo)

Built using TinyGo. To avoid conflicts with the project's `go.mod`, it is recommended to run this from `/tmp` or use the Makefile:

```bash
cd /tmp

tinygo build \
  -o /path/to/xynon/examples/plugins/add-header/add-header.wasm \
  -target wasip1 \
  /path/to/xynon/examples/plugins/add-header/main.go
```

Or using the Makefile:

```bash
make wasm
```

## Run

```bash
./bin/xynon -config examples/config.yaml
```

### Flags

| Flag | Default | Description |
|------|---------|-------------|
| `-config` | `config.yaml` | Path to the configuration file |
| `-no-hot-reload` | `false` | Disable hot-reloading and keep the initial plugin chain fixed |

## Plugin Management CLI

These commands manage plugins without starting the proxy server.

### plugin list — Show Plugins

```bash
./bin/xynon plugin list -config examples/config.yaml
```

Displays plugins in the configured directory and their registration order in the chain.

### plugin add — Add Plugin

```bash
./bin/xynon plugin add -config examples/config.yaml ./path/to/my-plugin.wasm
```

Copies the specified plugin file to the plugin directory and appends it to the chain in the configuration file. If a plugin with the same name exists, it overwrites the file.

### plugin remove — Remove Plugin

```bash
./bin/xynon plugin remove -config examples/config.yaml my-plugin
```

Removes the specified plugin from the configuration chain. The plugin file itself is not deleted.

### plugin build — Build TinyGo Plugin

```bash
./bin/xynon plugin build -config examples/config.yaml ./examples/plugins/add-header
```

Builds TinyGo source code from the specified directory and outputs `<directory-name>.wasm` to the plugin directory in the configuration.

*   `TINYGO`: Path to the tinygo binary (resolved from PATH if unset).
*   `XYNON_TINYGO_GOROOT`: GOROOT passed to TinyGo (specify if using a non-default Go environment).

## Admin CLI

### admin status — Show Backend Status

```bash
./bin/xynon admin status
```

Displays the health check status (Healthy / Unhealthy) of backend servers configured in upstreams.

## Configuration File (YAML)

```yaml
listen: ":8080"
allow_local_network: true
metrics:
  enabled: true
  address: ":9090"

plugins:
  dir: "./examples/plugins"
  chain:
    - name: auth            # defaults to type: wasm
    - name: rate-limit
      type: wasm
      config:
        capacity: 5
        refill_rate: 1
    - name: my-rpc-plugin
      type: rpc
      path: "/path/to/rpc-plugin-binary"

upstreams:
  - name: my-backend-cluster
    algorithm: round_robin
    servers:
      - url: http://localhost:8081
    circuit_breaker:
      enabled: true
      error_threshold: 5
      timeout: "30s"
    health_check:
      active:
        enabled: true
        path: "/health"
        expected_status: 200
        interval: "10s"
        timeout: "2s"
        max_fails: 3
      passive:
        enabled: true
        max_fails: 5
        fail_timeout: "30s"
```

| Field | Required | Description |
|-------|----------|-------------|
| `listen` | ✓ | Listening address (e.g., `:8080`) |
| `allow_local_network` | - | Allow proxying to local network addresses |
| `metrics` | - | Prometheus metrics configuration (`enabled`, `address`) |
| `plugins.dir` | ✓ | Directory where WASM files are stored |
| `plugins.chain[]` | - | Plugins to enable and their execution order. Supports `type` (wasm/rpc), `config`, and `path`. |
| `upstreams` | - | Backend clusters for load balancing, health checks, and circuit breaking |

## Hot-Reloading in Action

1. Start the proxy:
   ```bash
   ./bin/xynon -config examples/config.yaml
   ```

2. Send a request from another terminal:
   ```bash
   curl -x http://localhost:8080 http://httpbin.org/headers
   ```
   Verify that the response includes headers modified by the plugins (e.g., `X-Xynon: true`).

3. Overwrite a plugin's WASM file (replace with a new version):
   ```bash
   cp new-plugin.wasm examples/plugins/add-header/add-header.wasm
   ```
   The proxy logs will show `hot-reload complete`, and the new version will be applied to subsequent requests.

4. Gracefully shut down the proxy using `Ctrl+C` (SIGINT) or SIGTERM.

## Plugin Development

Xynon supports two plugin types: WASM (via wazero) and RPC (via hashicorp/go-plugin).

### WASM Plugins (ABI v1)

WASM plugins must implement the ABI defined in `internal/plugin/abi/abi.go`.

**Required Exported Functions:**

```go
//export xynon_abi_version
func xynon_abi_version() int32 { return 1 }

//export on_request
func on_request(handle int32) int32 { /* 0=continue, 1=short-circuit, 2=error */ }

//export on_response
func on_response(handle int32) int32 { /* 0=continue, 1=short-circuit, 2=error */ }
```

**Host Functions (env namespace):**

| Function | Signature | Description |
|----------|-----------|-------------|
| `get_header` | `(handle, namePtr, nameLen, outPtr, outCap) -> i32` | Read a header value |
| `set_header` | `(handle, namePtr, nameLen, valPtr, valLen) -> i32` | Set a header value |
| `set_status` | `(handle, status) -> i32` | Set response status code |
| `short_circuit` | `(handle, status) -> i32` | Respond immediately with status (valid only in on_request) |
| `xynon_log` | `(handle, msgPtr, msgLen) -> i32` | Write to proxy log |

Handles are only valid during the duration of each hook call. Using them afterwards returns an error.

### RPC Plugins

RPC plugins are standard Go binaries that implement the `hashicorp/go-plugin` interface defined in `internal/plugin/rpc_plugin.go`. They run out-of-process and communicate over gRPC.

## ABI Versioning Policy

- The ABI version is managed by the `CurrentVersion` constant in `internal/plugin/abi/abi.go`.
- The host only loads plugins that match its supported version. Unsupported versions are rejected at load time.
- **Breaking changes** (changing or removing function signatures) require incrementing `CurrentVersion`.
- **Backwards-compatible additions** (adding new host functions) can be done within the same version. Plugins are not required to import unused host functions.
