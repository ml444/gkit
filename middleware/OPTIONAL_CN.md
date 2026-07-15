# 可选依赖说明

部分中间件能力放在**可选子模块**中，避免主模块 `gkit` 强依赖 Redis / Prometheus / OTel。

## Prometheus 指标

使用 `-tags prometheus` 编译：

```bash
go get github.com/prometheus/client_golang/prometheus
go build -tags prometheus ./...
```

无 tag 时使用空实现（`metrics/prometheus_stub.go`）。

## Redis 分布式限流

```go
import (
    "github.com/ml444/gkit/middleware/ratelimit"
    rlredis "github.com/ml444/gkit/middleware/ratelimit/redis"
)

store := rlredis.NewStore(client, rlredis.Config{Service: "user-api"})
ratelimit.FrequencyLimitWithStore(store, cfgs, ratelimit.WithServiceName("user-api"))
```

Redis key 格式：

```text
gkit:rl:{service}:{path}:{windowMs}
```

## OpenTelemetry

- 轻量 Trace ID：`middleware/tracing`（解析 W3C `traceparent` 与 `X-Trace-Id`，无 OTel SDK）
- 完整 Span：`middleware/tracing/otel` + `pkg/tracing`

```go
import "github.com/ml444/gkit/middleware/tracing/otel"

httpx.SetHTTPMiddlewares(
    otel.HTTPMiddleware(),
    logging.HTTPMiddleware(),
    metrics.HTTPMiddleware(),
)
```

**勿**在同一服务上同时启用 `tracing.HTTPMiddleware()` 与 `otel.HTTPMiddleware()`。tracing/otel 须在 logging、metrics 之前。

## CSRF

面向 **Cookie/Session Web**。默认 `SkipBearer: true`，带 `Authorization: Bearer` 的请求**不校验 CSRF**（纯 Bearer API 无需启用）。浏览器表单场景使用 `csrf.DefaultOptions()`。
