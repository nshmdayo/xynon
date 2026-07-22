# Metrics Specification

## Overview

Xynon exposes Prometheus metrics to provide observability into proxy traffic. Metrics are exposed via a `/metrics` endpoint on a configured address.

## Configuration

Metrics can be enabled in `config.yaml` using the `metrics` section.

```yaml
listen: ":8080"
metrics:
  enabled: true
  address: ":9090"
plugins:
  dir: "./examples/plugins"
  chain: []
```

- `enabled`: If true, the metrics server starts.
- `address`: The address the metrics server listens on. Defaults to `:9090`.

## Exposed Metrics

The following metrics are exported:

| Metric Name | Type | Labels | Description |
|---|---|---|---|
| `xynon_requests_total` | Counter | `status`, `method`, `path` | Total number of HTTP requests processed. |
| `xynon_request_duration_seconds` | Histogram | `status`, `method`, `path` | Request duration in seconds. |

## Path Normalization

To prevent cardinality explosion in Prometheus labels, request paths are normalized.
- Numeric path segments (e.g., `/api/users/123`) are replaced with `{id}`.
- UUID path segments (e.g., `/api/users/550e8400-e29b-41d4-a716-446655440000`) are replaced with `{id}`.

