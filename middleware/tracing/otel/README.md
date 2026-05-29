# middleware/tracing/otel

OpenTelemetry **span** middleware for gkit. Depends on [`pkg/tracing`](../../../pkg/tracing/).

## Minimal integration

```go
// main.go
shutdown, _ := tracing.Setup(tracing.Config{
    ServiceName:  "my-svc",
    OTLPEndpoint: os.Getenv("OTEL_EXPORTER_OTLP_ENDPOINT"), // or "localhost:4318"
    OTLPInsecure: true,
})
defer shutdown(context.Background())

// server
import otelmw "github.com/ml444/gkit/middleware/tracing/otel"

srv := httpx.NewServer(
    httpx.SetHTTPMiddlewares(otelmw.HTTPMiddleware()),
    httpx.Middleware(otelmw.Server()),
)
```

## Without OTel backend

Use [`middleware/tracing`](../) only (trace ID in headers/logs, no Jaeger):

```go
httpx.SetHTTPMiddlewares(tracing.HTTPMiddleware())
httpx.Middleware(tracing.Server())
```

No `pkg/tracing` import required.
