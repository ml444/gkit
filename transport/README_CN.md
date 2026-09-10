# transport 使用与迁移说明

`httpx` 和 `grpcx` 提供 HTTP/gRPC 服务端、客户端、请求上下文、中间件及服务发现。HTTP 还提供 JSON、XML、Protobuf、表单和二进制编解码。以下说明对应本次 T01—T12 修复后的行为。

## HTTP 客户端地址与 HTTPS

```go
client, err := httpx.NewClient(
    httpx.WithEndpoint("https://api.example.com/v1"),
    httpx.WithTimeout(3*time.Second),
)
if err != nil { return err }
defer client.Close()
err = client.Invoke(ctx, http.MethodGet, "/users?limit=10", nil, &reply)
// 请求地址：https://api.example.com/v1/users?limit=10
```

完整 URL 的 scheme 决定协议。HTTPS 默认执行 TLS 证书校验，不需要为了启用 HTTPS 而额外提供 TLSConfig；自定义信任根、客户端证书等才使用 `WithTLSConfig`。裸 `host:port` 沿用默认 HTTP，有 TLSConfig 时默认 HTTPS。使用自定义 RoundTripper 时，其底层传输能力由调用方负责。

endpoint 不接受账号密码、查询参数或 fragment；查询参数放在每次调用路径中。配置错误尽量在 `NewClient` 返回。`NewClient()` 不提供 endpoint 时可用于 `Do(req)`，此时保留请求自己的完整 URL；`Invoke` 则要求 endpoint 或服务发现目标。

配置了 endpoint 的 `Do(req)` 会以其作为目标并拼接基础路径。Do 使用请求副本，不修改调用方的 URL、Host 和 Header；响应体仍由调用方关闭。Invoke 内部负责关闭响应体。

```go
client, err := httpx.NewClient(
    httpx.WithEndpoint("discovery:///user-service"),
    httpx.WithDiscovery(discoveryClient, ""),
)
```

HTTP 服务发现支持 IPv4 和 IPv6。显式服务名也可以通过 `WithDiscovery(discoveryClient, "user-service")` 提供。

## 路由分组与中间件

```go
srv := httpx.NewServer()
root := srv.GetRouter()
root.GET("/public", publicHandler)

admin := srv.NewRouteGroup("/admin", authHTTPMiddleware)
admin.HandleFunc("/users", usersHandler) // /admin/users，执行鉴权
admin.Group("/v1").GET("/settings", settingsHandler)
admin.Use(auditHTTPMiddleware) // 仅影响 admin 及其子组，包括之前注册的路由
```

`GET`、`POST`、`Handle`、`HandleFunc`、`HandlePrefix`、`HandleHeader` 都应用所属分组前缀和中间件。根路由 `Use` 作用于匹配路由；分组 `Use` 只作用于本组及子组。分组 Header 路由不会匹配具有相似名字的其他路径，例如 `/admin` 组不会匹配 `/administrator`。

执行顺序为全局、外层分组、内层分组、路由级，返回时反向执行。所有路由与中间件配置都必须在接收请求前完成；不支持并发热修改路由。404/405 等未匹配到业务路由的行为遵循底层 mux 的规则，不能假定路由中间件一定执行。

存在两种不同的中间件接口：

| 类型 | 配置入口与适用范围 |
| --- | --- |
| `middleware.HttpMiddleware` | `SetHTTPMiddlewares`、Router.Use、Group 参数或路由参数，包装 HTTP handler。手写和生成 handler 均可使用。 |
| `middleware.Middleware` | HTTP 的 `SetMiddlewares`/`Middleware` 保存 Service middleware，生成的 handler 负责调用；普通手写 HTTP handler 不会自动调用这条链。gRPC 的 `Middlewares` 默认应用于 unary RPC。 |

手写 HTTP handler 如需复用 Service middleware，应显式构造服务处理链，并使用 `httpx.NewCtx(w, r)` 的绑定及返回方法；鉴权等直接面向 HTTP 的操作可以使用 HTTP middleware。

```go
handler := middleware.Chain(serviceMiddlewares...)(
    func(ctx context.Context, req interface{}) (interface{}, error) {
        return service.Call(ctx, req.(*Request))
    },
)
// 在 HTTP handler 内完成 Bind 后：
out, err := handler(httpx.NewCtx(w, r), &in)
httpx.NewCtx(w, r).Returns(out, err)
```

gRPC 流式 RPC 只有显式配置 `EnableStreamMiddleware()` 才执行 Service middleware。它在流建立时执行一次，request 为 nil，不是每条消息执行一次；中间件必须能处理这种调用方式。

## 路径参数与编码错误

推荐使用可以返回错误的新接口：

```go
path, err := httpx.EncodeURLWithError("/files/{id}", request, true)
if err != nil { return err }
err = client.Invoke(ctx, http.MethodGet, path, nil, &reply)
```

普通占位参数作为一个路径段进行转义，例如 `a/b` 变成 `a%2Fb`，`?`、`#` 不会变成查询或 fragment。支持普通字段及嵌套字段名称，Protobuf 模板可使用 proto 字段名或 JSON 字段名。缺失、空值或非单值路径参数返回错误；此轮不扩展 `{name=...}` 等复杂模板语法。

服务端如需接收参数内的编码斜杠，使用 `httpx.RouterUseEncodedPath()`，再通过 `Context.BindVars` 或 `Context.Vars` 读取解码一次的参数；不要在业务层再次解码。

旧 `EncodeURL` 保留签名并使用安全转义，但无法返回编码错误。新生成的 HTTP 客户端使用 `EncodeURLWithError`；因此重新生成后需要同时使用包含该接口的 gkit 版本。已有生成代码仍可编译。

表单 bytes 默认输出标准 Base64，输入同时兼容标准及 URL-safe Base64。FieldMask 编码不修改原始消息。未知开放枚举保留数值；无法处理的 map 值等编码失败返回错误，不返回被静默截断的数据。

## 请求头与响应头

HTTP 客户端中间件通过 `Transport.In()` 添加、替换或删除请求头，修改会在最终发起请求前同步：

```go
func withToken(next middleware.ServiceHandler) middleware.ServiceHandler {
    return func(ctx context.Context, req interface{}) (interface{}, error) {
        if tr, ok := transport.FromContext(ctx); ok {
            tr.In().Set("Authorization", "Bearer "+token)
            tr.In().Delete("X-Internal")
        }
        return next(ctx, req)
    }
}
```

每次 Invoke 拥有独立 Header。业务代码仍须避免在调用进行期间并发修改传入的 map、请求消息或中间件的共享状态。

服务端可通过 `Transport.Out()` 返回多值响应头：

```go
tr, _ := transport.FromContext(r.Context())
tr.Out().Append("Set-Cookie", "a=1; Path=/", "b=2; Path=/")
httpx.NewCtx(w, r).Result(http.StatusOK, reply)
```

Out 中的同名头覆盖已有值列表，未设置的响应头保留。多值不会被截断。务必在首次 WriteHeader、Write 或 Flush 前设置响应头。

204、205 和 HEAD 正常响应不会进行 body 解码，传入的 reply 保持原值。4xx/5xx 即使 body 为空仍返回错误。其他 JSON 响应仍要求有效内容。

默认响应编码器区分编码失败和提交后的写入失败。前者返回正常错误响应；后者保留原始错误并由 Context.Result 记录日志，不尝试写第二份错误 JSON。自定义 ResponseEncoder 仍需自行明确其提交与错误处理行为。

## gRPC 服务发现与超时

```go
client, err := grpcx.NewClient(
    grpcx.WithEndpoint("discovery:///user-service"),
    grpcx.WithDiscovery(discoveryClient, ""),
    grpcx.WithDialTimeout(5*time.Second),
    grpcx.WithCallTimeout(2*time.Second),
)
if err != nil { return err }
defer client.Close()
```

每个客户端绑定自己的 resolver 与实例反馈缓存。不同注册中心即使服务名称相同，也不会共享解析结果。RPC 的 peer 或实例反馈不可用时跳过反馈，保留原始 RPC 结果。调用方的 `grpc.Peer` CallOption 仍然有效。

| 配置 | 行为 |
| --- | --- |
| `grpcx.WithTimeout(d)` | 同时设置建连和 unary 默认调用超时，保留旧入口；0 关闭两者。 |
| `grpcx.WithDialTimeout(d)` | 单独设置建连超时，0 关闭。默认建连非阻塞；如需等待连接就绪，另传 `grpc.WithBlock()`。 |
| `grpcx.WithCallTimeout(d)` | 单独设置 unary 调用超时，0 关闭；调用方更短的 deadline 优先。流式 RPC 只使用调用方 context。 |
| `grpcx.Timeout(d)` | 服务端 unary handler 的 context deadline；handler 及其下游应遵守取消信号。 |
| `httpx.WithTimeout(d)` | HTTP 客户端请求超时；0 关闭。 |
| `httpx.Timeout(d)` | HTTP 服务端请求 context deadline；不是强制终止 handler，也不自动发送超时响应。 |
| `httpx.ReadTimeout` / `ReadHeaderTimeout` / `WriteTimeout` / `IdleTimeout` | HTTP 服务端的连接读写与空闲超时，区别于业务 context 超时。 |

gRPC 客户端默认建连与 unary 调用超时均为 10 秒。选项按传入顺序应用，后面的选项覆盖对应值；负超时在客户端构造阶段报错。修复前未受默认超时限制的长 unary 调用，需显式调整 `WithCallTimeout`。

直接使用 resolver 时，推荐为每个客户端创建 `resolver.NewBuilder(dc)` 并传入 `grpc.WithResolvers(builder)`。旧的包级 Register 与地址查询仅保留兼容用途；Register 固定第一个全局注册中心，不能用于多注册中心隔离。

## 生命周期与迁移检查

保留现有接口：HTTP 使用 `Start(ctx)`，gRPC 使用 `Start()`，二者都通过 `Stop(ctx)` 关闭。Endpoint 方法可能提前创建监听器，不是纯查询。无需继续使用的实例即使尚未 Start，也应 Stop 释放提前创建的监听器。

gRPC 支持 Start 前 Stop、启动失败后 Stop 和重复 Stop；Stop 的 context 超时后强制停止。不承诺同一个 Server 在 Stop 后重新 Start。

升级时检查：

- 将原先依赖分组 Handle 注册到根路径的调用移到根路由；将原先通过分组 Use 配置的全局 middleware 移到根 Use。
- 确认完整 endpoint 的基础路径没有在每次调用路径中重复书写。
- 检查长 unary RPC 的超时设置，并区分建连与调用超时。
- 检查原来把字段中的 `/` 当作多级路径的调用；普通参数现在保持为单个路径段。
- 若对 Base64 字符串做签名或逐字节比较，更新对默认输出格式的预期。
- 重新生成 HTTP 客户端时同步升级传输库，避免新生成代码调用旧版本没有的接口。

## 验证入口

```sh
go test -count=1 ./transport/...
go test -race -count=1 ./transport/...
```

HTTP 生成器和生成示例是独立 Go module，需要在各自目录单独测试。测试使用本地临时监听端口、httptest TLS 服务和 gRPC bufconn；不需要真实注册中心。本次改动没有包含真实 xDS 控制面联调和压力测试。
