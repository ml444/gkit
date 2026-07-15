# middleware

gkit 的中间件层，面向 **HTTP** 与 **类 gRPC 的 Service 处理器**，统一处理日志、认证、限流、响应包装等横切逻辑，供 `httpx`、`grpcx` 及 `protoc-gen-go-http` 生成代码共用。

## 核心概念

### 两类中间件

| 类型 | 签名 | 使用场景 |
|------|------|----------|
| **Service 中间件** | `middleware.Middleware`，包装 `ServiceHandler` | `httpx` 生成 Handler、`grpcx` 一元 RPC、`httpx.Client` |
| **HTTP 中间件** | `middleware.HttpMiddleware`，包装 `http.Handler` | `httpx.Server` 路由（全局或路由组） |

定义见 [`middleware.go`](middleware.go)：

```go
type ServiceHandler func(ctx context.Context, req interface{}) (rsp interface{}, err error)
type Middleware       func(ServiceHandler) ServiceHandler
type HttpMiddleware   func(http.Handler) http.Handler
```

组合使用 `middleware.Chain(...)`、`middleware.HTTPChain(...)`。**第一个参数是最外层**（请求进入时最先执行）。

### Lurker 钩子

[`option.go`](option.go) 中的 `LurkerFunc` 用于 Handler 执行后的副作用（`logging`、`recovery`）：

```go
type LurkerFunc func(ctx context.Context, p interface{}) (err error)
```

`LurkerChain` 遇错即停；`ForceLurkerChain` 执行全部并忽略错误。

### Transport 上下文

多数 Service 中间件通过 `transport.FromContext(ctx)` 读取 **路径**（HTTP 路由模板或 gRPC full method）、**元数据** 等。`httpx.Server`、`grpcx.Server` 默认会注入（除非开启 `DisableTransportCtx`）。

---

## 如何注册

### httpx 服务端

```go
import (
    "time"

    "github.com/ml444/gkit/middleware/cors"
    "github.com/ml444/gkit/middleware/logging"
    "github.com/ml444/gkit/middleware/ratelimit"
    "github.com/ml444/gkit/middleware/recovery"
    "github.com/ml444/gkit/middleware/requestid"
    "github.com/ml444/gkit/middleware/response"
    "github.com/ml444/gkit/middleware/tracing"
    "github.com/ml444/gkit/middleware/validate"
    "github.com/ml444/gkit/transport/httpx"
)

srv := httpx.NewServer(
    httpx.Address(":5050"),
    // HTTP 层（路由）
    httpx.SetHTTPMiddlewares(
        requestid.HTTPMiddleware(),
        tracing.HTTPMiddleware(),
        cors.Default(),
        response.WrapHttpResponse(),
    ),
    // Service 层（每个生成 Handler）
    httpx.Middleware(
        recovery.Recovery(),
        tracing.Server(),
        requestid.Server(),
        ratelimit.FrequencyLimit(/* ... */),
        logging.LogRequest(),
        response.WrapError(),
        validate.Validator(),
    ),
)
```

单路由 HTTP 中间件：`router.GET("/path", handler, cors.New(opts))`。

### grpcx 服务端

```go
grpcx.NewServer(
    grpcx.Middlewares(
        recovery.Recovery(),
        tracing.Server(),
        ratelimit.FrequencyLimit(/* ... */),
        validate.Validator(),
        response.WrapError(),
    ),
)
```

服务端默认注册 `response.ServerErrorInterceptor` 做 gRPC 错误转换；panic 可用 `recovery.UnaryServerInterceptor()`。

### protoc-gen-go-http 生成 Handler

```go
user.GetUser_HTTP_Handler(srv, recovery.Recovery(), validate.Validator())
```

---

## 推荐顺序

**Service（外 → 内）：**

```
recovery → tracing → requestid → ratelimit → auth → circuitbreaker → timeout → validate
  → [业务 Handler] → logging → response.WrapError → response.ReplaceEmptyResponse → response.WrapResponse
```

**HTTP（外 → 内）：**

```
requestid → tracing → security → cors → ipfilter → csrf → gzip → logging.HTTPMiddleware → response.WrapHttpResponse
```

**recovery** 放在 Service 链最外层，以便捕获内层 panic。**response.WrapHttpResponse** 放在 HTTP 链末尾，用于统一 JSON 响应体 `{code, message, data}`。

---

## 各包说明

### `response` — 错误、空响应、响应包装

| API | 层级 | 说明 |
|-----|------|------|
| `WrapError()` | Service | 将 error 转为 `errorx.Error` |
| `ReplaceEmptyResponse(data)` | Service | nil / 零值指针响应替换为 `data` |
| `WrapResponse()` | Service | 成功 proto 响应包装为 `ApiCommonResponse` |
| `WrapHttpResponse()` | HTTP | 2xx JSON 包装为 `{code, message, data}` |
| `MarkHttpRaw(w)` / `IsHttpRaw(h)` | HTTP | 标记原始响应，跳过包装 |
| `ServerErrorInterceptor` | gRPC | 服务端错误 → gRPC status + details |
| `ClientErrorInterceptor` | gRPC | 客户端还原 `errorx.Error` |

```go
httpx.Middleware(
    response.WrapError(),
    response.ReplaceEmptyResponse(&emptypb.Empty{}),
    response.WrapResponse(),
)
```

文件下载等二进制接口在 Handler 内调用 `response.MarkHttpRaw(w)`，避免 `WrapHttpResponse` 修改 body。

---

### `recovery` — Panic 恢复

| API | 层级 | 说明 |
|-----|------|------|
| `Recovery(fns...)` | Service | 捕获 panic、打印栈；默认返回 `errorx.InternalServer` |
| `UnaryServerInterceptor(fns...)` | gRPC | 一元 RPC panic 恢复 |
| `StreamServerInterceptor(fns...)` | gRPC | 流式 RPC panic 恢复 |

自定义示例：

```go
recovery.Recovery(func(ctx context.Context, stack interface{}) error {
    log.Errorf("panic: %v\n%v", ctx.Value(recovery.RecoverKey{}), stack)
    return errorx.InternalServer("服务内部错误")
})
```

上下文键：`recovery.RecoverKey`、`recovery.RequestKey`。

---

### `ratelimit` — 频率限流

按 **transport path** 限流（HTTP 路由模板或 gRPC 方法名）。支持精确路径、正则、全局（`MatchKindAll`），每条规则可配置多个时间窗口。

| API | 说明 |
|-----|------|
| `FrequencyLimit(cfgs...)` | 进程内固定窗口（默认） |
| `FrequencyLimitWithOptions(cfgs, opts...)` | 支持 `WithFailClosed`、`WithStore` |
| `FrequencyLimitWithStore(store, cfgs, opts...)` | 基于 `Store` 的分布式限流 |
| `NewMemoryStore()` | 内存 `Store` 实现 |

```go
ratelimit.FrequencyLimit(
    &ratelimit.LimitCfg{
        Kind:  ratelimit.MatchKindAll,
        Freqs: []*ratelimit.Frequency{{Period: time.Second, Limit: 100}},
    },
    &ratelimit.LimitCfg{
        Kind:  ratelimit.MatchKindExact,
        Paths: []string{"/api/v1/users/{id}"},
        Freqs: []*ratelimit.Frequency{{Period: time.Minute, Limit: 60}},
    },
)
```

| 选项 | 作用 |
|------|------|
| `WithFailClosed()` | 无 transport 上下文时拒绝（默认放行） |
| `WithStore(store)` | 配合 `FrequencyLimitWithOptions` 使用自定义存储 |
| `WithServiceName(name)` | Redis key 中的 `{service}` 段 |

分布式限流使用可选子模块 `middleware/ratelimit/redis`（根模块不依赖 go-redis）：

```go
import rlredis "github.com/ml444/gkit/middleware/ratelimit/redis"

store := rlredis.NewStore(client, rlredis.Config{Service: "user-api"})
ratelimit.FrequencyLimitWithStore(store, cfgs, ratelimit.WithServiceName("user-api"))
```

Redis key 格式：`gkit:rl:{service}:{path}:{windowMs}`。详见 [OPTIONAL.md](OPTIONAL.md)。

超限返回 `ratelimit.ErrLimitExceed`（HTTP 429）。

---

### `validate` — 参数校验

| API | 说明 |
|-----|------|
| `Validator()` | 若 req/rsp 实现 `validate.IValidator`，则调用 `Validate()` |

配合 [**protoc-gen-go-validate**](https://github.com/ml444/gkit/tree/master/cmd/protoc-gen-go-validate)。校验失败会规范为 `errorx` 错误。

```go
httpx.Middleware(validate.Validator())
```

---

### `logging` — Service 与 HTTP 日志

| API | 层级 | 说明 |
|-----|------|------|
| `LogRequest(fns...)` | Service | Handler 结束后记录耗时、路径、Trace ID、请求体（可自定义 `LurkerFunc`） |
| `HTTPMiddleware()` | HTTP | 结构化访问日志：method、path、status、字节数、耗时、trace、span（从 request context 读取） |

Service 日志上下文键：`logging.Took`（毫秒）、`logging.Reply`（响应）。

```go
httpx.SetHTTPMiddlewares(logging.HTTPMiddleware())
httpx.Middleware(logging.LogRequest())
```

---

### `tracing` — 链路 Trace ID

| API | 层级 | 说明 |
|-----|------|------|
| `Server()` | Service | 保证 context 中有 trace ID，写入出站 metadata |
| `HTTPMiddleware()` | HTTP | 解析 W3C `traceparent` 或 `X-Trace-Id`，注入 context 并回写响应头 |
| `UnaryServerInterceptor()` | gRPC | 一元 RPC trace |

请求头定义见 [`pkg/header`](../pkg/header/header.go)（含 W3C `traceparent`）。**中间件顺序：** `tracing`（或 `otel`）须在 `logging`、`metrics` 之前；同一服务勿同时启用 `tracing.HTTPMiddleware()` 与 `otel.HTTPMiddleware()`。完整 OpenTelemetry 见 [`pkg/tracing`](../pkg/tracing/)（独立子模块）。

---

### `requestid` — 请求 ID

| API | 层级 | 说明 |
|-----|------|------|
| `HTTPMiddleware()` | HTTP | 传递或生成 `X-Request-ID` |
| `Server()` | Service | 保证 context 中有 request ID |
| `FromContext(ctx)` | — | 读取 request ID |

---

### `metrics` — 指标

| API | 层级 | 说明 |
|-----|------|------|
| `HTTPMiddleware()` | HTTP | 按 method/path 计数与耗时；有 trace ID 时 histogram 附加 exemplar |
| `Server()` | Service | 按 transport path 记录 |
| `SetRecorder(r)` | — | 注入自定义 `Recorder` |
| `NewInMemoryRecorder()` | — | 测试用 |

**编译标签 `prometheus`：** 见 [OPTIONAL.md](OPTIONAL.md)。

```bash
go build -tags prometheus ./...
```

OpenTelemetry：轻量 trace ID 见 `middleware/tracing`；完整 Span 见 `middleware/tracing/otel`。

---

### `auth` — 认证

| API | 层级 | 说明 |
|-----|------|------|
| `Server(opts...)` | Service | 从 transport metadata 取 token 校验 |
| `HTTPMiddleware(opts...)` | HTTP | `Authorization: Bearer` 或 API Key |
| `FromContext(ctx)` | — | 读取 `Claims` |
| `StaticValidator(map)` | — | 静态 token 表 |

```go
auth.HTTPMiddleware(
    auth.WithValidator(auth.StaticValidator(map[string]auth.Claims{
        "my-token": {"user_id": "1"},
    })),
    auth.WithSkipPaths("/health"),
)
```

错误：`auth.ErrUnauthorized`（401）、`auth.ErrForbidden`（403）。JWT 等可实现 `auth.TokenValidator`。

---

### `security` — 安全响应头

| API | 层级 | 说明 |
|-----|------|------|
| `HTTPMiddleware(opt)` | HTTP | HSTS、防嗅探、X-Frame-Options 等 |
| `DefaultOptions()` | — | 推荐默认配置 |

---

### `cors` — 跨域

| API | 说明 |
|-----|------|
| `Default()` | 开发环境宽松配置 |
| `New(Options)` | Origin 白名单、方法、头、Credentials、MaxAge、预检 |

```go
cors.New(cors.Options{
    AllowOrigins:     []string{"https://app.example.com"},
    AllowMethods:     "GET,POST,PUT,DELETE,OPTIONS",
    AllowHeaders:     "Content-Type,Authorization",
    AllowCredentials: true,
    MaxAge:           3600,
})
```

带 `Access-Control-Request-Method` 的 `OPTIONS` 预检直接返回 `204`，不进入业务 Handler。

---

### `csrf` — CSRF 防护

| API | 层级 | 说明 |
|-----|------|------|
| `HTTPMiddleware(opt)` | HTTP | Cookie + `X-CSRF-Token` 双提交校验 |

适用于 Cookie/Session Web。**默认 `SkipBearer: true`**：带 `Authorization: Bearer` 的请求跳过 CSRF（纯 Bearer API 无需 CSRF）。使用 `csrf.DefaultOptions()`；若需强制校验可设 `SkipBearer: false`。

```go
csrf.HTTPMiddleware(csrf.DefaultOptions())
```

---

### `ipfilter` — IP 黑白名单

| API | 层级 | 说明 |
|-----|------|------|
| `HTTPMiddleware(opt)` | HTTP | CIDR 允许/拒绝 |
| `ParseCIDRs(...)` | — | 解析 CIDR |

支持 `TrustXFF` 解析 `X-Forwarded-For`。

---

### `gzip` — 响应压缩

| API | 层级 | 说明 |
|-----|------|------|
| `HTTPMiddleware(opt)` | HTTP | 客户端支持 gzip 时压缩响应 |

`MinLength` 可设最小压缩长度阈值。

---

### `timeout` — 单次请求超时

| API | 层级 | 说明 |
|-----|------|------|
| `HTTPMiddleware(d)` | HTTP | 为 request context 设置超时 |
| `Server(d)` | Service | 单次 RPC 超时 |

与 `httpx.Timeout`（全服超时）互补；同时使用时常把 `timeout.Server` 放在更靠近 Handler 的内层。

---

### `bodylimit` — 请求体大小（Service）

| API | 说明 |
|-----|------|
| `Server(maxBytes)` | 近似判断请求体过大并拒绝 |

HTTP 侧请配合 `httpx.MaxRequestBodySize`。

---

### `circuitbreaker` — 熔断

| API | 说明 |
|-----|------|
| `Server(opt)` | 按 path 熔断：关闭 → 打开 → 半开 |

```go
circuitbreaker.Server(circuitbreaker.Options{
    Threshold:    5,              // 连续失败次数
    OpenDuration: 30 * time.Second,
})
```

熔断打开时返回 `circuitbreaker.ErrOpen`（503）。

---

### `retry` — 客户端重试

| API | 说明 |
|-----|------|
| `Middleware(opt)` | 可配置次数、退避、重试条件 |
| `Client(opt, next)` | 包装单个 `ServiceHandler` |

```go
httpx.NewClient(
    httpx.WithMiddlewares(retry.Middleware(retry.Options{
        MaxAttempts: 3,
        Backoff:     100 * time.Millisecond,
    })),
)
```

---

### `idempotency` — 幂等

| API | 层级 | 说明 |
|-----|------|------|
| `HTTPMiddleware(store, ttl)` | HTTP | 校验 `Idempotency-Key` 头 |
| `Server(store, ttl, keyFn)` | Service | 从 context 自定义 key |
| `NewMemoryStore()` | — | 进程内存储 |

重复请求返回 `idempotency.ErrDuplicate`（409）。

---

## 编译标签与可选模块

详见 [OPTIONAL.md](OPTIONAL.md)。

---

## 目录结构

```
middleware/
├── middleware.go      # Middleware、HttpMiddleware、Chain
├── option.go          # LurkerFunc
├── logging/           # Service 日志 + HTTP 访问日志
├── bodylimit/         # 请求体限制
├── circuitbreaker/    # 熔断
├── cors/              # 跨域
├── csrf/              # CSRF
├── gzip/              # 压缩
├── idempotency/       # 幂等
├── ipfilter/          # IP 过滤
├── logging/           # Service 日志
├── metrics/           # 指标
├── ratelimit/         # 限流
├── recovery/          # Panic 恢复
├── requestid/         # 请求 ID
├── response/          # 响应/错误包装
├── retry/             # 客户端重试
├── security/          # 安全头
├── timeout/           # 超时
├── tracing/           # Trace ID
└── validate/          # 校验
```

---

## 相关文档

- 项目总览：[README.md](../README.md) / [README_CN.md](../README_CN.md)
- 传输层：[transport/httpx](../transport/httpx)、[transport/grpcx](../transport/grpcx)
- 错误模型：[errorx](../errorx)
