# Dev-dependencies

Everything you need for the development process

- OpenAPI [linter](https://github.com/daveshanley/vacuum)

## Metrics pipeline

Metric collection lives outside the HTTP layer: collectors run once with the
application context, no matter how many clients are connected.

```
PostgreSQL ──► collectors (internal/service/metrics) ──► hub (internal/infra/hub) ──► SSE handler (internal/transport/http)
                              │
                              └──► history storage (Badger, internal/service/metrics/repo/storage) ──► REST (internal/transport/http)
```

| Component | Responsibility |
| --- | --- |
| `internal/model/metrics` | metric names, typed `Number` values, measurements, history entities |
| `internal/service/metrics` | collectors with their own intervals, publish + persist, retention cleanup |
| `internal/infra/hub` | fan-out to SSE subscribers, bounded queues, drop-oldest overflow policy |
| `internal/transport/http/metrics.go` | SSE endpoint, batching window, heartbeat; history REST handler |
| `internal/service/metrics/repo/storage` | Badger time series (`metric:{name}:{timestamp}`) |

### Endpoints

- `GET /api/metrics/stream` — one SSE connection per client for all metrics:

  ```text
  retry: 3000

  event: metrics
  data: {"buffer_cache_hit":{"value":98.4,"timestamp":"2026-10-11T18:00:01Z"},"connections":{"value":26,"timestamp":"2026-10-11T18:00:00Z"}}
  ```

  Each measurement carries its own timestamp, only changed metrics are sent, and
  updates that arrive inside `METRICS_BATCH_WINDOW` are merged into one event
  (the newest sample of a metric wins). A `metrics` event with the latest known
  value of every metric is sent right after subscribing, so a client can combine
  it with `GET /api/metrics/history` without gaps or duplicates: fetch the
  history up to the moment of subscribing and apply live samples only when their
  timestamp is newer than the last stored point.

  The route is registered next to the generated OpenAPI router
  (`transport/http.MetricsStreamPath`), because a strict OpenAPI handler must
  return a complete body while the stream is endless.

- `GET /api/metrics/history?names=&from=&to=&limit=` — initial series and
  recovery of points missed while a client was offline.

### Configuration

All settings are optional environment variables (see `internal/config/metrics.go`).

| Variable | Default | Meaning |
| --- | --- | --- |
| `METRICS_BUFFER_CACHE_HIT_INTERVAL` | `1s` | shared buffers cache hit rate collection interval |
| `METRICS_CONNECTIONS_INTERVAL` | `3s` | connections collection interval |
| `METRICS_DATABASE_SIZE_INTERVAL` | `30s` | database size collection interval |
| `METRICS_QUERY_TIMEOUT` | `3s` | timeout of a single collection (including SQL) |
| `METRICS_BATCH_WINDOW` | `50ms` | SSE batching window |
| `METRICS_BATCH_SIZE` | `64` | maximum number of metrics per SSE event |
| `METRICS_SUBSCRIBER_BUFFER` | `32` | per-subscriber queue size |
| `METRICS_HEARTBEAT_INTERVAL` | `15s` | SSE comment interval |
| `METRICS_RETRY_INTERVAL` | `3s` | reconnect delay advertised to the browser |
| `METRICS_HISTORY_RETENTION` | `1h` | how long history points are kept |
| `METRICS_HISTORY_CLEANUP_INTERVAL` | `1m` | retention cleanup period |
| `METRICS_HISTORY_LIMIT` | `600` | default points per metric in the history response |

### Adding a metric

1. Add the name to `metrics.Names()` in `internal/model/metrics`.
2. Add a collection function in `internal/service/metrics` that returns
   `[]metrics.Update` (use one query for several indicators when possible, and
   take the timestamp from the database when the query provides it).
3. Register the collector with its interval in `Service.collectors()`.
4. Extend the frontend client if the metric must be displayed.

Errors inside one collector are logged and never stop the other collectors;
while the connection manager reports "not connected", collections are skipped.
