# protoc-gen-go-http

Generate HTTP routes, handlers, and clients from `google.api.http` annotations. The plugin targets [gkit `httpx`](../../transport/httpx) and supports the **pluck** extension for file upload/download scenarios.

## Install

```shell
go install github.com/ml444/gkit/cmd/protoc-gen-go-http@latest
```

## Quick start

Your proto must import `google/api/annotations.proto` and include `pluck/` on the proto path when using pluck options:

```shell
protoc \
  --proto_path=. \
  --proto_path=$GOMODCACHE/github.com/grpc-ecosystem/grpc-gateway@v1.16.0/third_party/googleapis \
  --go_out=paths=source_relative:. \
  --go-http_out=paths=source_relative:. \
  ./api/user.proto
```

This generates `user_http.pb.go` with:

- `RegisterUserHTTPServer` — register routes on `*httpx.Server`
- `UserHTTPServer` — interface to implement
- `UserHTTPClient` — typed HTTP client (optional, see flags below)

## Plugin options

Pass options via `--go-http_out`:

| Option | Default | Description |
|--------|---------|-------------|
| `omitempty=true` | `true` | Skip generation when no RPC has `google.api.http` |
| `omitempty=false` | — | Generate default `POST /{service}/{method}` routes for RPCs without HTTP rules |
| `omitempty_prefix=/api` | `""` | Prefix for default paths when `omitempty=false` |
| `module=github.com/foo/bar` | `github.com/ml444/gkit` | Base import path for `httpx`, `middleware`, and `pluck` in generated code |
| `client=full` | `full` | Generate HTTP client (`client=none` for server-only) |
| `warnings=warn` | `warn` | Emit stderr warnings for suspicious HTTP rules (`off` or `error`) |

Examples:

```shell
# Generate server handlers only
protoc --go-http_out=paths=source_relative,client=none:. user.proto

# Fail CI on HTTP rule warnings
protoc --go-http_out=paths=source_relative,warnings=error:. user.proto

# Custom module path for a forked gkit
protoc --go-http_out=paths=source_relative,module=github.com/acme/gkit:. user.proto
```

> **Note:** With the default `omitempty=true`, protos without any `google.api.http` annotation produce **no** `*_http.pb.go` file. Use `omitempty=false` if you want fallback POST routes.

## Pluck extension

Import `pluck/pluck.proto` and add method options for header/body mapping:

| Scenario | HTTP rule | Pluck option |
|----------|-----------|--------------|
| JSON upload | `body: "*"` | — |
| Raw bytes field upload | `body: "file_data"` | `pluck.request.headers_to` |
| Raw body stream upload | `body: "*"` | `pluck.request.body_to` |
| JSON download | `body: "*"` | — |
| Raw bytes download | `response_body: "data"` | optional `pluck.response.headers_from` |
| File download with headers | `body: "*"` | `pluck.response.body_from` + `headers_from` |

See [tests/storage/storage.proto](tests/storage/storage.proto) for working examples and [tests/run.go](tests/run.go) for an integration demo.

## Generated handler naming

When a method has `additional_bindings`, each binding gets a unique handler suffix (`UploadV00`, `UploadV01`, …). The HTTP client generates one method per binding; duplicate RPC names use suffixes (`UploadV0_1`) when multiple bindings exist.

## Route prefix registration

Use `RegisterXxxHTTPServerWithPrefix` to mount all routes under a prefix:

```go
RegisterUserHTTPServerWithPrefix(s, "/api/v1", srv) // => /api/v1/users/...
RegisterUserHTTPServer(s, srv)                      // equivalent to prefix ""
```

## Raw responses and unified JSON wrapper

Handlers for `response_body` / pluck download routes are annotated with `// +gkit:http_raw` and call `response.MarkHttpRaw(w)` so [`WrapHttpResponse`](../../middleware/response/wrap_response.go) skips JSON wrapping for binary responses.

## Streaming RPCs

Streaming RPCs with `google.api.http` annotations are rejected at codegen time with a clear error.

## Development

```shell
go test ./...
```

The golden test in `generator_test.go` compares output against `tests/storage/storage_http.pb.go` (requires `protoc` on `PATH`).

Regenerate the fixture:

```shell
ANNOT=$(go env GOMODCACHE)/github.com/grpc-ecosystem/grpc-gateway@v1.16.0/third_party/googleapis
protoc --proto_path=. --proto_path="$ANNOT" \
  --go-http_out=paths=source_relative:. \
  --go_out=paths=source_relative:. \
  ./tests/storage/storage.proto
```
