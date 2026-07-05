# Product Requirements Document: RPC Plugins (`HandlerRPCClient` / `HandlerRPCServer`)

## Feature Description
Out-of-process plugins executed via standard RPC to ensure process isolation.

## Requirements
- **Runtime**: `hashicorp/go-plugin` over standard `net/rpc`.
- **Execution Model**: Out-of-process. The proxy host launches the RPC plugin as a standalone child process.
- **Data Passing**: To avoid complex shared-memory management across processes, HTTP headers and statuses are passed by value during the RPC call. The guest mutates the copy and returns the updated state to the host.
- **Constraints**: RPC plugins can use standard Go, but must implement the expected `hashicorp/go-plugin` interfaces.
