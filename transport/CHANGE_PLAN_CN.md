# transport 模块改动方案（已确认）

日期：2026-09-10。

状态：用户已确认实施 T01—T12、第 4 节及第 6 节；第 5 节可选增强暂缓。实施与验收结果见文末。

提案阶段仅新增本文档。工作区原有 `transport/httpx/coding.go` 未提交修改保持原样，后续实施时应在其当前内容上处理，不能直接覆盖或还原。

## 1. 目标与建议实施范围

目标是修复已确认的功能缺陷，统一容易误用的接口行为，并用回归测试约束这些行为。

建议本轮实施以下范围：

- 第 3 节全部缺陷修复，共 12 项；其中部分条目包含同一功能下的多个缺陷。
- 第 4 节的低风险易用性改进：配置校验、文档、示例和必要的兼容接口。
- 第 6 节的回归测试与关联调用适配。

第 5 节为后续可选增强，不包含在上述建议范围内。确认实施建议范围，不代表授权所有可选增强。

本轮不升级 Go、gRPC 或 Protobuf 依赖，不重写整个传输层，不引入自动重试，不改造真实 xDS 控制面，也不直接更改现有公共接口的函数签名。

## 2. 分析依据与现状

已检查 `transport` 实现、现有测试、相关中间件及 HTTP 生成客户端/服务端调用方式。

前次执行 `go test ./transport/...` 未通过，主要结果如下：

- gRPC 服务发现调用返回 `grpc peer is nil`。
- gRPC 停止逻辑发生空指针 panic。
- resolver 的一个测试直接构造未注入 DiscoveryClient 的 builder，测试用例与当前实现不一致；需与产品缺陷分别处理。
- HTTP 相关包当次报告通过，其中使用了测试缓存，不能据此认定边界行为正确。

另外使用仓库外的 18 个临时测试复现了 URL、路由、元数据、Protobuf、超时、生命周期及响应错误问题。这些测试按期望行为断言，当前失败是缺陷证据，尚未作为正式测试加入仓库。

P1 表示应优先修复的调用失败、鉴权边界或 panic 问题；P2 表示其余功能正确性问题。以下均说明触发条件，不表示所有请求都会出现相应故障。

## 3. 缺陷修复内容

### T01：统一 HTTP 路由分组行为（P1）

涉及：`httpx/router.go`、相关路由测试。

当前问题：

- `Group("/admin", auth).HandleFunc("/secret", handler)` 注册到 `/secret`，绕过分组前缀及分组中间件。
- `Handle`、`HandlePrefix`、`HandleHeader` 同样没有完整应用分组语义。
- 分组调用 `Use` 会作用到共享底层 router，影响分组外的公共路由。

拟改动：所有注册入口统一应用完整分组前缀和中间件。根路由的 `Use` 保持全局语义，分组的 `Use` 仅作用于该组及子组。中间件顺序明确为全局、外层分组、内层分组、路由级，退出时反向执行；应用应在开始接收请求前完成配置。

验收：受保护接口只能通过预期的分组路径访问；鉴权中间件执行且只执行一次；公共路由不受其他分组的 `Use` 影响；嵌套组、尾斜杠、方法匹配、`WalkRoute` 和路径模板保持正确。

兼容性：依赖“分组 Handle 注册到根路径”或“分组 Use 实际全局生效”的调用需要迁移到根路由显式注册。修复前的行为不继续保留为默认行为。

### T02：修复 HTTP endpoint、HTTPS 与地址拼接（P1）

涉及：`httpx/client.go`、`httpx/option.go`、客户端测试。

当前问题：完整 endpoint 被直接赋值到 `URL.Host`，基础路径丢失；是否提供 TLSConfig 被错误地用于决定所有请求的协议。

拟改动：

- endpoint 在构造阶段解析并校验，后续使用解析后的 scheme、authority 和基础路径。
- `WithEndpoint("https://example.com/api")` 配合调用路径 `/users`，得到 `https://example.com/api/users`。
- 显式 `http://` 或 `https://` 决定协议。HTTPS 未提供自定义 TLSConfig 时使用默认 TLS 校验，不降级为 HTTP。
- 无 scheme 的 `host:port` 保留现有默认选择：未提供 TLSConfig 时使用 HTTP，提供时使用 HTTPS。
- 未指定 endpoint 的 `Do(req)` 保留请求自己的 scheme 和地址；`Invoke` 缺少可用目标时返回清晰错误。
- discovery 地址使用 `net.JoinHostPort` 拼接，以支持 IPv6；保留服务发现的目标替换能力。
- 对不支持的 scheme、缺失服务名及含歧义的 endpoint 返回配置错误。首轮不支持在 endpoint 内携带账号密码、查询参数或 fragment，调用级查询仍放在请求路径中。
- 请求与 URL 使用独立副本，避免修改调用方持有的请求对象。

验收：裸地址、HTTP/HTTPS URL、带基础路径、查询参数、IPv6、服务发现分别覆盖；HTTPS 测试验证实际 TLS 请求及证书校验。基础路径与特殊字符不能被重复转义或错误清理。

兼容性：完整 URL 与 HTTPS 改为按其声明含义执行；自定义 RoundTripper 仍由调用方负责底层传输，包装层不得擅自改写其请求协议。

### T03：修复 gRPC 实例反馈对 RPC 结果的覆盖（P1）

涉及：`grpcx/client.go`、客户端及服务发现测试。

当前问题：调用方 context 通常不包含实际服务端 peer；反馈逻辑因此把正常 RPC 结果替换成 `grpc peer is nil`。

拟改动：通过 `grpc.Peer` CallOption 获取本次 RPC 的实际 peer，并保留调用方已有 CallOption 的行为。缺少 peer 或找不到实例时跳过反馈并记录诊断信息，始终保留原始 RPC 结果及错误。成功/失败反馈保持按实际实例归属，不把业务错误直接当作连接故障。

验收：成功 RPC 不再被反馈失败覆盖；原始 Unavailable、DeadlineExceeded 和业务错误保留；调用方自行传入 `grpc.Peer` 时仍可获取地址；无 peer 的失败分支不产生二次错误。

### T04：隔离 gRPC resolver 与实例缓存（P1）

涉及：`grpcx/resolver/discovery.go`、`grpcx/client.go`、resolver 测试。

当前问题：全局 `sync.Once` 绑定第一个 DiscoveryClient；每个服务刷新会覆盖全局地址缓存。

拟改动：

- 每个 grpcx 客户端创建绑定自身 DiscoveryClient 的 resolver builder，通过连接级 resolver 配置使用，不依赖进程全局注册决定归属。
- 实例缓存归属到客户端及其服务，不再用单一地址 map 保存所有服务。
- 实例下线、地址变化及 resolver 关闭时更新或清理自身状态；一个 resolver 不得清空另一个的缓存。
- 瞬时注册中心错误保留上一份可用地址，服务明确无实例时清空对应服务地址。
- 串行化同一 resolver 的刷新，关闭时取消在途发现请求，避免关闭后继续更新连接。
- 现有导出 `Register`、地址缓存查询和测试辅助函数保留签名作为兼容入口，并说明其限制；grpcx 新实现不再依赖全局入口。此轮不直接删除公共符号。

验收：两个客户端使用不同注册中心且服务名相同，仍各自连接正确实例；同一注册中心的两个服务不相互覆盖；刷新、下线、短暂失败、重复关闭及并发读取均正确。

兼容性：保留源码兼容，但推荐调用方使用 `grpcx.WithDiscovery`，避免依赖全局注册的初始化顺序。

### T05：修复 Protobuf 表单编解码 panic（P1）

涉及：`httpx/coder/form/proto_encode.go`、`proto_decode.go`、form 测试。

拟改动：

- map 的 key 使用 `MapKey()` 描述符，value 使用 `MapValue()`；遍历中的编码错误传递给上层，不返回不完整数据加 nil error。
- 未知枚举值不再解引用空描述符。开放枚举保留数值往返能力；不支持的值返回明确错误。
- Timestamp/Duration 输入 `null` 按字段未设置处理，避免构造或解引用空消息；其他无法解析的值返回带字段信息的错误。

验收：`map<string,bool>`、不同 key/value 类型、未知枚举、Timestamp/Duration 的正常值、null 和非法输入都有覆盖；请求输入不得导致编解码器 panic。

### T06：区分 gRPC 建连超时与调用超时（P2）

涉及：`grpcx/client.go`、`grpcx/client_option.go`、客户端测试。

当前问题：`WithTimeout` 注释声明默认调用超时，实际只作用于 DialContext。

拟改动：新增 `WithDialTimeout` 和 `WithCallTimeout`；保留 `WithTimeout` 签名，明确其作为便捷选项同时设置两个默认值，单项选项按应用顺序覆盖对应值。默认值保持现有 10 秒配置，并让调用超时实际作用于 unary RPC。值为 0 表示关闭对应的默认超时，负值返回配置错误。

调用方 context 的更短 deadline 优先；建连继续保持现有默认非阻塞行为，建连超时不代表等待服务就绪。流式 RPC 不自动套用 unary 默认超时，由调用方 context 控制生命周期。

验收：慢 unary RPC 在默认调用超时内结束；更短的调用方 deadline 生效；0 可关闭默认超时；阻塞建连和非阻塞建连行为分别验证；流式调用不受新增 unary 超时误伤。

兼容性：此前长 RPC 因缺陷未超时，修复后可能开始超时。迁移说明应给出显式延长或关闭调用超时的方法。

### T07：修复 gRPC 提前停止及失败清理（P2）

涉及：`grpcx/server.go`、生命周期测试。

拟改动：仅在 listener 非空时关闭；支持未 Start 就 Stop、Start 失败后 Stop、重复 Stop，以及 Endpoint 已创建监听器但尚未 Serve 的资源回收。优雅停止超时后继续强制停止并返回 context 错误。

验收：上述场景均不 panic；提前创建的端口被释放；正在处理的请求按正常优雅停止或超时强制停止语义结束。

兼容性：不在此轮承诺同一个 Server 实例 Stop 后可重新 Start，也不修改现有 Start/Stop 签名。

### T08：修复 HTTP 请求与响应元数据丢失（P2）

涉及：`httpx/client.go`、`httpx/transport.go`、`httpx/response_writer.go`、`httpx/ctx.go`。

拟改动：

- 明确客户端 `Transport.In()` 为可修改的出站请求头，在最终发起请求前同步中间件的修改，覆盖添加、替换及删除操作。
- 每次调用拥有独立元数据，避免污染调用方传入的 Header 或并发请求。
- 输出响应头时保留全部值，统一 Context 与 ResponseWriter 的写入路径，避免重复或仅保留第一个值。
- `Out()` 中设置的同名响应头覆盖已有值列表，未设置的响应头保持原样；响应提交后的修改不再尝试补写已发送的头。

验收：中间件注入 Authorization 等头能到达服务端；删除头生效；多个 Set-Cookie 全部返回；并发调用之间无串值或重复响应头。

### T09：保证 bytes 与 FieldMask 编解码一致（P2）

涉及：`httpx/coder/form/proto_encode.go`、`proto_decode.go`、form 测试。

拟改动：bytes 默认输出标准 Base64，解码兼容现有 URL-safe 格式；FieldMask 的 snake_case 到 camelCase 转换在副本上完成，不修改调用方消息。

验收：包含 `0xfb, 0xff` 的 bytes 能编码后解码还原；标准及 URL-safe 编码均能接收；FieldMask 编码前后原对象完全相同，连续或并发只读编码不发生数据修改。

兼容性：接受旧输入格式，但默认输出格式修正为标准 Base64；如调用方对编码字符串做签名或逐字节比较，需要同步更新预期。

### T10：正确转义 URL 路径参数并提供错误返回（P2）

涉及：`httpx/client.go`、URL 测试；必要时关联 `cmd/protoc-gen-go-http` 的客户端生成逻辑和测试样例。

拟改动：新增 `EncodeURLWithError(pathTemplate, msg, needQuery) (string, error)`。普通占位参数作为单个路径段转义，`?`、`#`、`/`、`%` 等不能改变 URL 结构；查询参数通过 URL 编码 API 拼接。新接口明确返回编码失败、缺少必要路径参数等错误。

现有 `EncodeURL` 保留原签名并复用安全转义逻辑；由于旧签名无法返回错误，保留兼容错误处理并标记推荐迁移。后续生成的 HTTP 客户端使用新接口并向调用方返回错误，已有生成代码继续可编译。此轮不扩展复杂 Google HTTP 路径模板语法。

验收：路径特殊字符、中文、嵌套字段、查询参数及缺失参数均覆盖；检查最终 HTTP 请求中的路径与查询含义，不只比较辅助函数返回的字符串。需要接收编码斜杠的服务端示例说明 `RouterUseEncodedPath` 配置。

兼容性：包含斜杠的普通字段值将保持一个路径段，不再隐式扩展为多级路径。

### T11：正确处理无响应体的 HTTP 结果（P2）

涉及：`httpx/coding.go`、客户端解码测试。

拟改动：204、205 以及 HEAD 的正常响应跳过响应体解码，reply 保持原值；普通声明有实体的 JSON 响应仍校验格式，不能把所有空 body 一律当作成功。4xx/5xx 即使无 body，也返回保留 HTTP 状态的错误。

验收：204/205/HEAD 在 reply 非空时不报 JSON EOF；普通非法 JSON 仍报错；无 body 的错误响应不变成成功。

### T12：保留响应写入失败且避免重复写错误响应（P2）

涉及：`httpx/coding.go`、`httpx/ctx.go`、响应写入测试。

当前工作区的未提交修改在 `w.Write` 失败时执行 `println` 后返回 nil。该改动可能意在避免重复写入错误响应，但目前同时隐藏了原始写入失败。

拟改动：区分“响应提交前的编码失败”和“响应提交后的写入失败”。默认编码器返回可识别且保留原始 cause 的写入错误；`Context.Result` 对此类错误使用项目日志记录并结束，不再调用 ErrorEncoder 写第二份响应。提交前的编码失败继续走正常错误响应。

验收：注入 Write 失败时编码器返回错误；日志可诊断原始错误；Context 不重复发送状态码或追加错误 JSON；正常响应内容与状态码保持不变。

兼容性：在现有未提交改动基础上实现其避免重复写入的目标，不机械还原文件。首轮沿用现有日志设施，不增加新的全局错误回调 API。

## 4. 本轮包含的易用性改进

| 内容 | 拟交付行为 |
| --- | --- |
| 配置校验 | 在能够返回 error 的客户端构造入口尽早校验目标地址、发现配置与超时参数；HTTP 无 endpoint 的通用 Do 用法仍允许。 |
| 中间件说明 | 区分 HTTP middleware 与 Service middleware，说明手写 handler、生成 handler、分组及流式 RPC 的接入方式；提供鉴权分组示例。 |
| 超时说明 | 解释建连超时、unary 调用超时、服务端 context 超时与 HTTP 读写超时之间的区别，明确 0 的含义。 |
| 生命周期说明 | 保留 HTTP/gRPC 现有方法签名，明确 Endpoint 可能创建监听器、Stop 的资源回收范围及实例不可默认重启。 |
| 编解码与元数据示例 | 展示路径编码、多值响应头、客户端中间件改头以及新旧 URL 编码接口迁移方式。 |

实施时新增模块使用说明，并更新相关公开 API 注释及必要的现有 README 示例；不只提供内部实现说明。

## 5. 后续可选增强：本轮默认不实施

这些条目涉及新的公共 API、策略选择或较大兼容性影响，单独列出便于后续确认。

| 编号 | 增强内容 | 建议方向与影响 |
| --- | --- | --- |
| U01 | 单项编解码定制 | 允许单独替换 BindBody、BindQuery、ResponseEncoder 等，无需实现整个 IRouterCoder。需确定选项组合与覆盖顺序。 |
| U02 | 统一 Context.JSON 行为 | 复用 JSON coder，使 Protobuf 与默认响应路径一致。可能改变枚举、64 位整数、字段名和默认值输出，需要迁移说明。 |
| U03 | 完整响应钩子 | 新增在解码前可观察状态与响应头的钩子，覆盖错误响应；保留原 OnResponse 语义。需明确 body 所有权及回调错误规则。 |
| U04 | 流式上传下载 | 提供不强制读完整 body 的调用接口，明确关闭响应体、取消、大小限制与超时责任。 |
| U05 | 协商与资源策略 | 正确解析 Accept 多类型及 q 权重，支持配置最大响应体、未知字段策略，以及可选的严格 406/415 行为；默认策略需兼容现有 JSON 回退。 |
| U06 | 更统一的服务生命周期接口 | 在保持旧接口可用的前提下增加适配层。是否提供无副作用的地址查询、统一 Start(ctx) 等，需要单独设计。 |

## 6. 文件范围与验收方式

预计主要修改：

- `transport/httpx` 下的 client、router、coding、ctx、response_writer、option、transport 及对应测试。
- `transport/httpx/coder/form` 下的 Protobuf 编解码及对应测试。
- `transport/grpcx` 下的 client、client_option、server、resolver 及对应测试。
- `transport` 模块使用说明、相关 API 注释和必要的根 README 示例。
- `cmd/protoc-gen-go-http`：仅在 T10 需要时适配新 URL 编码接口，并更新相关生成测试样例；不扩展本次无关的生成器功能。

实施顺序建议：

1. 将前次复现转换成正式、可重复的回归测试，修正 resolver 空 builder 的测试构造。
2. 修复 T01—T05，优先解决路由隔离、协议、服务发现及 panic。
3. 修复 T06—T12，完成超时、生命周期、元数据及编解码一致性。
4. 完成关联生成器适配、使用说明与迁移说明。
5. 执行验证，汇总实际文件清单、通过结果、兼容性变化和剩余限制供复核。

必要验证：

- `go test -count=1 ./transport/...`，避免仅依赖旧测试缓存。
- 对本次改动的 resolver、客户端及元数据相关包运行 race 检查，重点验证并发隔离和关闭行为。
- 使用 httptest/TLS 测试服务和 gRPC bufconn 验证实际请求行为，不只断言内部字段。
- 若修改 HTTP 生成器，分别执行其独立 Go module 的测试及相关生成客户端编译/集成测试；根模块测试不会自动覆盖所有嵌套 module。
- 校验中间件调用顺序、公开接口兼容性及错误保真；测试失败不得通过放宽断言掩盖。

本轮验收不包含真实注册中心故障演练、真实 xDS 控制面联调或吞吐压测；如需纳入，可扩充验收范围。

## 7. 确认后的执行边界

建议确认内容：实施第 3 节 T01—T12、第 4 节和第 6 节；第 5 节暂缓。

需接受的主要行为变化：分组隔离恢复正常、HTTPS 按 URL 执行、默认 unary 调用超时开始生效、路径参数正确转义、Base64 输出格式统一，以及写入失败不再返回成功。

如只希望先处理 P1，可将实施范围缩小为 T01—T05 及其测试、必要文档。若需要可选增强，请按 U01—U06 指定。

本文档本身不触发代码修改、提交、合并或发布。用户确认后按确认范围实施；提交、合并或发布仍以届时的明确要求为准。


## 8. 实施与验收记录

已按用户确认范围实施 T01—T12、第 4 节及第 6 节；U01—U06 未实施。没有升级依赖，没有提交、合并或发布。

已完成的内容：

- HTTP 分组的全部注册入口统一前缀与中间件，并保持分组隔离、嵌套顺序和尾斜杠语义。
- HTTP endpoint、HTTPS、基础路径、IPv6 地址拼接、请求副本及出站请求头同步修复。
- gRPC 使用连接级 resolver，按客户端和服务隔离实例反馈；初始化异步执行，关闭时取消在途刷新并清理实例状态。
- gRPC 原始错误与调用方 Peer CallOption 保留；建连和 unary 调用超时可分别设置，流式 RPC 不套用 unary 默认超时。
- 服务提前停止、启动失败清理、重复停止及监听器资源回收修复。
- Protobuf map、未知枚举、null 时间字段、bytes 和 FieldMask 的错误处理与往返行为修复。
- 新增 EncodeURLWithError，适配 HTTP 生成器及生成样例；服务端在启用编码路径时只解码路径变量一次。
- 多值响应头、无响应体响应、写入失败保真与避免重复错误响应修复。
- 新增 `transport/README_CN.md` 使用与迁移说明，更新相关根 README、生成器 README 和 API 注释。

已通过验证：

- `go test -race -count=1 ./transport/...`：完整模块回归及 race 检查通过。
- HTTP 生成器独立模块 `go test -count=1 ./...` 通过。
- `TestGoldenStorageHTTP` 使用真实 protoc 执行并通过，没有因缺少工具而跳过。
- 生成示例独立模块的 HTTP 往返集成测试及 race 检查通过，验证生成客户端、生成 handler、Service middleware 和两个 Set-Cookie。
- `git diff --check` 通过。

生成示例模块的原有依赖清单与根模块不一致，直接使用 `-mod=readonly` 会要求更新 go.mod。此次使用仓库外的临时 modfile 完成测试，仓库所有 go.mod/go.sum 保持不变；这项历史清单问题未作为本次修复扩展范围。

仍需注意的迁移变化：默认 unary RPC 超时现在实际生效，普通路径参数中的斜杠将被转义，表单 bytes 默认输出标准 Base64，以及旧全局 resolver 注册入口不提供多注册中心隔离。具体配置示例见模块使用说明。


额外执行根模块 `go test -count=1 ./...` 时，transport、discovery 及多数其他包通过，但全量命令未通过，失败位于本次未改动的以下模块：

- `config`：环境变量默认值、带等号的默认值和简写默认值测试失败。
- `middleware/csrf`：默认跳过 Bearer API 的测试实际返回 403。
- `optx`：字符串及若干整数范围 handler 的空值/指针处理测试失败。

这些模块未纳入本次修复范围，没有为通过全量测试而修改它们或放宽断言。此次修复的最终 HTTP 中间件链使用一次构造保留 handler 状态，相关 HTTP 和 transport 根包的 race 检查再次通过。
