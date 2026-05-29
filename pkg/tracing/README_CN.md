# pkg/tracing

面向 **gkit** 的 OpenTelemetry 追踪工具包。本目录为**可选子模块**（`github.com/ml444/gkit/pkg/tracing`），主模块默认不依赖 OTel SDK。

## 选型

| 需求 | 使用 |
|------|------|
| 仅日志/响应头中的 Trace ID | [`middleware/tracing`](../../middleware/tracing/) |
| 完整 Span、Jaeger 等后端 | **本包** + [`middleware/tracing/otel`](../../middleware/tracing/otel/) |

## gkit 快速接入

### 1. 启动时初始化（`main`）

```go
shutdown, err := tracing.Setup(tracing.Config{
    ServiceName:  "user-api",
    OTLPEndpoint: "localhost:4318", // OTLP/HTTP（Jaeger 1.35+、OTel Collector、Tempo）
    OTLPInsecure: true,
    SampleRatio:  0.1,
})
if err != nil {
    log.Fatal(err)
}
defer func() { _ = shutdown(context.Background()) }()
```

`InitTracer` 已标记废弃，请使用带 `shutdown` 的 `Setup`。

### 2. 挂载中间件

```go
import traceotel "github.com/ml444/gkit/middleware/tracing/otel"

httpx.NewServer(
    httpx.SetHTTPMiddlewares(traceotel.HTTPMiddleware()),
    httpx.Middleware(traceotel.Server()),
)

grpcx.NewServer(
    grpcx.Middlewares(traceotel.Server()),
    grpcx.UnaryInterceptor(traceotel.UnaryServerInterceptor()),
)
```

### 3. `go.mod` replace

```go
replace (
    github.com/ml444/gkit => ../gkit
    github.com/ml444/gkit/pkg/tracing => ../gkit/pkg/tracing
    github.com/ml444/gkit/middleware/tracing/otel => ../gkit/middleware/tracing/otel
)
```

## 核心 API

| API | 说明 |
|-----|------|
| `Setup(Config)` | 注册全局 TracerProvider，返回 `shutdown` |
| `NewTracer` + `Start`/`End` | 创建 Server/Client Span |
| `TraceID(ctx)` / `SpanID(ctx)` | 从 Span 读取 ID |
| `SyncTraceIDToCache` / `ClearTraceIDCache` | 与旧版按 goroutine 取 trace 的日志库兼容 |

## 优化点与现状

### 已改进

- 去掉 cache / gRPC stats 中的调试 `fmt.Println`
- 新增 `Setup`：可配置采样率、返回 `shutdown`
- OTel 中间件同步 `pkg/header` 与可选 goroutine cache

### 建议后续优化

| 问题 | 说明 |
|------|------|
| 默认 OTLP/HTTP | 不再使用 Jaeger exporter；Jaeger 1.35+ 监听 4318（OTLP） |
| `Metadata` 传播器为空实现 | 需业务元数据时可补全 `Inject`/`Extract` |
| `SetClientSpan`/`SetServerSpan` 字段未赋值 | 属性不完整，目前主要依赖 `End` 记录 proto 大小 |
| `traceIdCache` 可能泄漏 | 新代码优先用 context；使用 cache 后需 `ClearTraceIDCache` |
| Go / semconv 版本不统一 | 子模块 go 1.19，semconv 混用 v1.4 与 v1.12 |

## 与 glog 集成（遗留）

若日志框架只能按 `routineId` 取 trace：

```go
cfg.TradeIDFunc = func(entry *message.Entry) string {
    return tracing.GetTraceIdFromCache(entry.RoutineId)
}
```

启用 OTel 中间件后会自动 `SyncTraceIDToCache`；请求结束 `ClearTraceIDCache`。

更推荐：`header.CorrelationID(ctx)` 或 `tracing.TraceID(ctx)`。

## 相关文档

- 英文说明：[README.md](README.md)
- 中间件可选依赖：[middleware/OPTIONAL_CN.md](../../middleware/OPTIONAL_CN.md)
