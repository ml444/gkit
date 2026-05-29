# Optional dependencies

Some middleware features live in **optional submodules** so the main `gkit` module stays lightweight.

## Prometheus metrics

Build with `-tags prometheus`:

```bash
go get github.com/prometheus/client_golang/prometheus
go build -tags prometheus ./...
```

Uses `middleware/metrics/prometheus.go`. Without the tag, `metrics` uses a no-op recorder (`prometheus_stub.go`).

## Redis distributed rate limiting

Import the submodule (does not pull go-redis into the root module):

```go
import (
    "github.com/ml444/gkit/middleware/ratelimit"
    rlredis "github.com/ml444/gkit/middleware/ratelimit/redis"
)

store := rlredis.NewStore(redisClient, rlredis.Config{Service: "user-api"})
mw := ratelimit.FrequencyLimitWithStore(
    store,
    cfgs,
    ratelimit.WithServiceName("user-api"),
)
```

Redis key format:

```text
gkit:rl:{service}:{path}:{windowMs}
```

Example: `gkit:rl:user-api:api/v1/users/{id}:60000`

## OpenTelemetry tracing

Lightweight trace ID propagation: `middleware/tracing` (no OTel SDK).

Full OTel spans: import `middleware/tracing/otel` (depends on `pkg/tracing`):

```go
import "github.com/ml444/gkit/middleware/tracing/otel"

httpx.SetHTTPMiddlewares(otel.HTTPMiddleware())
httpx.Middleware(otel.Server())
```

## CSRF

CSRF (`middleware/csrf`) targets **cookie/session web apps**. By default `SkipBearer: true` skips CSRF when `Authorization: Bearer` is present — **pure Bearer APIs do not need CSRF**. Use `csrf.DefaultOptions()` or set `SkipBearer: false` to enforce CSRF on all clients.
