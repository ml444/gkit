# transport 使用与迁移说明

`httpx` 和 `grpcx` 提供 HTTP/gRPC 服务端、客户端、请求上下文、中间件及服务发现。HTTP 还提供 JSON、XML、Protobuf、表单和二进制编解码。以下说明覆盖 T01—T12 修复、U01—U04，以及 U05 的配置基础和响应解码大小配置。

## 单项编解码与实例 JSON 配置

通过 `NewRouterCoder` 只替换需要定制的部分，再用现有 `RouterCoder` 选项传给服务器：

```go
rc, err := httpx.NewRouterCoder(
    httpx.WithBindBody(customBodyDecoder),
    httpx.WithErrorEncoder(customErrorEncoder),
)
if err != nil { return err }
srv := httpx.NewServer(httpx.RouterCoder(rc))
```

可单独设置 `WithBindVars`、`WithBindQuery`、`WithBindForm`、`WithBindBody`、`WithResponseEncoder`、`WithErrorEncoder`。不传的项保留默认实现；选项按从左到右执行，同一项后者覆盖前者。nil 选项、nil 函数及 typed nil 的基础实现或 JSON coder 都在 `NewRouterCoder` 返回错误。

`WithBaseRouterCoder(base)` 一次替换六项回调，通常放在单项覆盖前；放在后面会覆盖之前设置的六项回调。基础实现必须提供完整的六项非 nil 函数。已有第三方 `IRouterCoder` 无需增加方法。

每个服务器可以使用自己的 Protobuf JSON 配置：

```go
// import jsoncodec "github.com/ml444/gkit/transport/httpx/coder/json"
opts := jsoncodec.DefaultOptions()
opts.Marshal.UseProtoNames = true
opts.Marshal.EmitUnpopulated = false
opts.Unmarshal.DiscardUnknown = false

rc, err := httpx.NewRouterCoder(
    httpx.WithJSONCoder(jsoncodec.NewCoder(opts)),
)
if err != nil { return err }
srv := httpx.NewServer(httpx.RouterCoder(rc))
```

`WithJSONCoder` 作用于本实例默认的请求体绑定、正常响应、错误响应和 JSON 回退。传入的 codec 必须非 nil 且 `Name()` 为 `json`。它不会自动切换 `Context.JSON` 的模式，也不影响 HTTP 客户端；这两项分别通过下文的 `WithJSONMode` 和 `WithDecodeJSONCoder` 设置。

自定义回调和 `WithBaseRouterCoder` 携带的回调按自身配置执行，不受外层 `WithJSONCoder` 强行改写。若需要给基础实现配置 JSON，应在构造基础实现时传入该选项。

配置边界如下：

| 入口 | 配置与兼容行为 |
| --- | --- |
| `NewServer()` 原默认路径 | 保留动态全局 codec 查询及原 JSON 配置行为 |
| `NewRouterCoder(...)` | 构造时复制注册表；内置 JSON coder 同时冻结当前全局 JSON 选项 |
| `jsoncodec.NewCoder(opts)` | 按值保存选项，不在请求阶段读取全局 JSON 变量；不隐式补默认值 |
| `jsoncodec.DefaultOptions()` | 复制当前兼容默认值，适合在此基础上覆盖单项配置 |
| `coder.LookupCoder(name)` | 忽略名称大小写的严格查询，未注册返回 nil,false，不回退 JSON |
| `coder.Snapshot()` | 返回独立注册表副本；修改 map 不影响全局注册表，codec 对象仍共享 |

`jsoncodec.Options{}` 使用 protojson 零值行为，与 `DefaultOptions()` 不同。普通 Go 对象仍使用标准库 JSON；自定义 Marshaler/Unmarshaler 保持优先。`DiscardUnknown` 只针对 Protobuf，普通 struct 的严格未知字段策略尚未加入。

全局注册表使用 `atomic.Pointer` 发布不可变 map 快照，Get/Lookup 不获取互斥锁；Snapshot 复制同一个已发布版本，返回独立 map。Register 使用写锁串行复制、修改并发布新版本，避免并发注册丢失更新。每次注册需要 O(N) 复制，适合 codec 数量少、读多写少的场景。旧 `GetCoder` 仍保留未匹配回退 JSON 的规则。

快照不保护自定义 codec 内部状态；自定义回调、codec 对象以及 protojson 选项引用的 Resolver 必须支持并发使用。

已构造的新 router coder 不受后续注册替换影响。用户自行注册的自定义 JSON coder 按原对象保留，不深拷贝或替换。旧的 `jsoncodec.MarshalOptions`、`UnmarshalOptions` 只应在启动初始化时设置，不支持运行中并发修改。

当前保留现有 Accept 选择方式和 10 MiB 默认响应解码上限。可通过下文的 decoder 工厂覆盖解码上限；Accept 新协商和 406/415 严格模式仍待后续实施。

## Context.JSON 的编码模式

默认 `JSONLegacy` 保留标准库 Encoder 的输出，包括末尾换行。需要与同实例的默认 JSON 响应使用相同 Protobuf 编码规则时，显式启用：

```go
rc, err := httpx.NewRouterCoder(
    httpx.WithJSONMode(httpx.JSONCodec),
    httpx.WithJSONCoder(jsoncodec.NewCoder(jsoncodec.DefaultOptions())),
)
if err != nil { return err }
srv := httpx.NewServer(httpx.RouterCoder(rc))
```

`Context.JSON` 明确输出 JSON，不根据 Accept 改为 XML，也不调用自定义业务 `ResponseEncoder`。新模式先编码再提交状态，编码失败时 handler 仍可处理错误；写入失败后返回带原因的错误，`ReturnError` 不会再输出第二个响应。

新模式下 `JSON(200, nil)` 输出 `null`，HEAD、204、205、304 不写正文。切换前应检查 Protobuf 枚举、int64/uint64、字段名、未设置字段和末尾换行的变化。旧 `IRouterCoder` 不用增加方法；可选 `JSONEncoderProvider` 用于自定义 JSON 方法。`Context.Reset` 会重新读取新请求绑定的 coder 并重置状态。

## 响应头钩子和响应解码配置

```go
decode, err := httpx.NewResponseDecoder(
    httpx.WithDecodeMaxBytes(32 << 20), // 成功/错误响应的解码读取上限
    httpx.WithDecodeJSONCoder(jsoncodec.NewCoder(jsoncodec.DefaultOptions())),
)
if err != nil { return err }
client, err := httpx.NewClient(
    httpx.WithEndpoint("https://api.example.com"),
    httpx.WithResponseDecoder(decode),
    httpx.WithResponseHeadersHook(func(ctx context.Context, meta httpx.ResponseMeta) error {
        // 解码前可读取 meta.StatusCode 和 meta.Header，包括 4xx/5xx。
        return nil
    }),
)
if err != nil { return err }
defer client.Close()

err = client.Invoke(ctx, http.MethodGet, "/users", nil, &reply,
    httpx.OnResponseHeaders(func(ctx context.Context, meta httpx.ResponseMeta) error {
        // 此次调用的钩子在客户端级钩子之后运行。
        return nil
    }),
)
```

新钩子按配置顺序追加，每个获得独立 Header 副本，没有 Body 读取权限。网络失败且没有可用响应时不执行；只观察重定向后的最终响应，不覆盖 1xx 或中间重定向响应。返回错误会终止后续钩子和解码，关闭响应体，通过 `ResponseHookError` 保留 HTTP 状态，支持 errors.Is/As；此时尚未解码业务错误正文。

原 `OnResponse` 仍在解码成功后运行，Download 则在复制成功后运行；最后配置的回调生效，nil 可禁用。此时 Body 可能已到 EOF，库负责关闭。钩子只用于 `Invoke`、`InvokeReader`、`Download`，原始 `Do` 保持调用方管理响应体的行为。

`WithDecodeMaxBytes` 默认 10 MiB，0 明确表示无限制，负数在构造时返回错误。对不需解码的成功响应（reply 为 nil、HEAD、204/205）仍跳过正文，不额外读完整 body 检查大小。`WithDecodeJSONCoder` 固定客户端 JSON 配置；其他 codec 也在工厂创建时形成快照。自定义 decoder 自行负责资源限制。

超限返回 `*ResponseTooLargeError`，含 Limit 和原 HTTP StatusCode；其 Unwrap 保留原 errorx 的 502/50201 分类。

服务发现反馈也已区分远端与本地失败：正常完成和已解码的 4xx 反馈成功，5xx/网络读取失败反馈失败；调用方取消或本地 hook/decoder/writer 错误不更新节点状态（已收到 5xx 的情况除外）。每次实际请求最多反馈一次，业务错误仍原样返回。

## 流式上传下载

大文件建议使用独立客户端关闭默认 2 秒总超时，再由每次调用的 context 控制期限：

```go
client, err := httpx.NewClient(
    httpx.WithEndpoint("https://files.example.com"),
    httpx.WithTimeout(0),
)
if err != nil { return err }
defer client.Close()
ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
defer cancel()

file, err := os.Open("input.bin")
if err != nil { return err }
err = client.InvokeReader(ctx, http.MethodPut, "/files/input.bin", file, &reply,
    httpx.SetRequestContentType("application/octet-stream"),
)
// InvokeReader 已接管 file 的关闭责任，验证参数失败也会关闭。
if err != nil { return err }

out, err := os.Create("output.part")
if err != nil { return err }
info, downloadErr := client.Download(ctx, "/files/result.bin", out,
    httpx.DownloadMaxBytes(1 << 30),
)
closeErr := out.Close() // 目标 writer 始终由调用方关闭。
if downloadErr != nil {
    // info.Bytes 是已经写入的字节数；output.part 可能包含部分内容。
    return downloadErr
}
if closeErr != nil { return closeErr }
_ = info
```

| 方法 | 正文处理与所有权 |
| --- | --- |
| `InvokeReader` | 不调用 request encoder、不整体缓存上传内容；传入 io.ReadCloser 时方法从入口接管关闭；响应走 decoder |
| `Download` | 发起 GET，用固定 32 KiB 缓冲区复制成功正文；关闭响应 Body，不关闭 dst，返回状态、响应头及已写入字节数 |
| `Do` | 原始 HTTP 操作，调用方读取并关闭响应 Body；适合自定义方法、Trailer、长连接等场景 |

Download 的成功流默认不限大小；`DownloadMaxBytes` 的 0 表示不限，负数在发送前报错，显式限额按实际读取数据计算。默认 gzip 自动解压时限制的是解压后字节数。超限探测最多额外读 1 字节，但不写入 dst；已声明的 Content-Length 超限可提前拒绝。

4xx/5xx 使用 decoder 的独立错误正文上限，错误页面不会写入 dst；最终未跟随的 3xx 返回 `UnexpectedStatusError`。限额、writer 错误、网络中断、取消或复制后的回调错误，均保留已写入的 Bytes。库不自动删除、重命名部分文件，不承诺原子替换。

两个便捷接口都复用请求头同步、中间件、响应钩子和服务发现。上传中间件输入为 reader，下载输入为 nil、输出为 DownloadInfo；只支持 Protobuf 输入的中间件需单独调整。流式调用拒绝中间件重复执行下游，避免重复写入或重读已消耗的源。库不新增自动重试机制；底层 net/http 重定向规则保持不变，已知 reader 的 ContentLength/GetBody 会保留，不可重放的 reader 不会被偷偷缓冲。

原客户端总超时仍覆盖正文传输。使用原始 Do 时应在 Body 读取/关闭后才取消 context；自定义阻塞 reader/writer 需自行配合取消。服务端继续用已有 Context.Stream，handler 管理输入 reader 的关闭，并按需求调整服务器的请求体大小、应用超时和读写超时。

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

推荐新代码使用统一适配层。HTTP 与 gRPC 都实现 `transport.ManagedServer`：

```go
// 先配置路由、中间件及服务，再交给适配器管理。
srv := httpx.NewServer(httpx.Address("127.0.0.1:0"))
srv.GetRouter().GET("/health", func(w http.ResponseWriter, r *http.Request) {
    w.WriteHeader(http.StatusOK)
})
managed, err := httpx.Managed(srv, transport.LifecycleOptions{
    ShutdownTimeout: 10 * time.Second,
})
if err != nil { return err }
defer managed.Stop(context.Background()) // 提前返回时也释放已绑定资源

if err := managed.Listen(); err != nil { return err }
endpoint, bound := managed.EndpointURL()
if bound {
    log.Printf("bound to %s", endpoint)
}

ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
defer cancel()
return managed.Start(ctx) // 阻塞；收到信号后等待优雅停止完成
```

上述片段放在返回 `error` 的应用启动函数内，`log` 为标准库包。无需提前取得地址时，可以直接 `Start(ctx)`，由它自动监听。gRPC 只需替换构造部分，后续使用方式相同：

```go
srv, err := grpcx.NewServer(grpcx.Address("127.0.0.1:0"))
if err != nil { return err }
// 在此注册生成的 gRPC 服务。
managed, err := grpcx.Managed(srv, transport.LifecycleOptions{})
if err != nil { return err }
```

| 操作 / 配置 | 行为 |
| --- | --- |
| `Listen()` | 显式绑定端口，绑定成功后重复调用幂等；单独 Listen 失败可重试。地址计算失败会释放本次新建的监听器。 |
| `EndpointURL()` | 无 I/O，只返回地址副本；调用者修改副本不影响服务器。尚未完成 Listen 时为 `nil, false`；停止期间及停止后 `ok=false`，可保留最后地址。 |
| `Start(ctx)` | 自动监听并阻塞到服务与关闭协调完成；调用前已取消则返回 `ctx.Err()`，不消耗启动机会。启动开始后只允许一次，启动失败也需要新建 Server。 |
| `Stop(ctx)` | 首次调用触发共享关闭，其后等待同一结果；ctx 仅限制当前调用者的等待。即使它已取消，也会启动关闭过程。 |
| `ShutdownTimeout` | 实际优雅关闭的期限，独立于 Start/Stop 的 context；0 使用 10 秒，负值构造失败。超时后 HTTP 强制 Close，gRPC 强制 Stop，返回可用 `errors.Is(err, context.DeadlineExceeded)` 判断的错误。 |

启动 context 取消且优雅停止成功时，`Start` 返回 nil。非预期 Serve 错误和关闭错误会保留，二者同时发生时合并返回。并发 Start 返回 `transport.ErrAlreadyStarted`；停止中的新 Start/Listen 返回 `transport.ErrServerStopped`。Stop 在从未启动、只监听过或已停止时也安全；注入的 Listener 同样由服务器负责关闭。

HTTP 请求保留启动 context 的值，但不继承其取消和 deadline，给在途请求留下优雅完成时间。`httpx.Timeout` 设置的业务超时仍然有效。gRPC 的在途 unary 和 stream RPC 等待完成，超过优雅期则断开连接。强制关闭不能终止忽略取消信号的业务 goroutine；HTTP hijack/WebSocket 连接、额外后台任务和 Shutdown 回调的完成，由应用自行协调。

HTTP 地址使用 http/https；显式 `httpx.Endpoint` 指定的发布 URL 保持原值。gRPC 在 TLSConfig 或已知 TLS Credentials 下返回 grpcs，否则返回 grpc；自定义或 xDS 动态凭据的实际安全模式仍由凭据配置决定。URL 是发布地址形式，不能直接假定为旧客户端接受的拨号字符串；例如 gRPC 直连可读取 `endpoint.Host`，并另外配置匹配的 TLS 凭据。

`EndpointURL` 的布尔值只表示适配器已经完成绑定，不代表已经接收请求或通过健康检查；注入或通过旧 Endpoint 提前创建的监听器，也要调用适配器 Listen 才发布快照。服务注册与注销仍由应用负责，监听成功本身不应当作为业务就绪的依据。

同一 Server 只能创建一个 Managed 适配器。交接后原 Start/Stop/Endpoint 返回 `transport.ErrLifecycleOwned`；也不要绕过适配器直接调用嵌入服务器的 Serve、Shutdown、Close、GracefulStop。已经调用过旧 Start/Stop 的实例不能再交接；只通过旧 Endpoint 绑定过的监听器可以交接。gRPC 原 RegisterDiscovery/DeregisterDiscovery 内部调用旧 Endpoint，属于旧生命周期用法；使用 Managed 时，请根据地址快照和应用就绪状态自行操作注册中心。

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
