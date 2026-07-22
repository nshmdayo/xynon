# Product Requirements Document: Proxy Engine

## Feature Description
The proxy engine is responsible for receiving incoming HTTP requests, executing the plugin chain, and forwarding requests to the upstream server.

## Requirements
- **Plugin Chain Execution**: Maintain a synchronized list of `plugin.Handler` interfaces. Execute the chain sequentially for each request (`OnRequest`) and in reverse order for each response (`OnResponse`).
- **Dynamic Load Balancing**: Route requests to upstream backend servers based on configuration (`round_robin`, `least_connections`, `ip_hash`).
- **Health Checks**: Support active (periodic polling) and passive (traffic monitoring) health checks for backend servers.
- **Admin Endpoints**: Provide `/_admin/upstreams` (restricted to localhost) to expose real-time backend health statuses in JSON.
- **Concurrent Safety**: Access to shared mutable states (e.g., the plugin Chain registry, upstreams, or configurations during hot reload) must be synchronized using Go's concurrent primitives (`sync.RWMutex`, `atomic.Pointer`, or channel-based synchronization).
