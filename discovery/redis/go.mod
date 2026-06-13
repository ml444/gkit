module github.com/ml444/gkit/discovery/redis

go 1.23.0

toolchain go1.23.4

replace github.com/ml444/gkit => ./../..

require (
	github.com/alicebob/miniredis/v2 v2.33.0
	github.com/ml444/gkit v0.0.0-00010101000000-000000000000
	github.com/redis/go-redis/v9 v9.14.0
)

require (
	github.com/alicebob/gopher-json v0.0.0-20200520072559-a9ecdc9d1d3a // indirect
	github.com/cespare/xxhash/v2 v2.3.0 // indirect
	github.com/dgryski/go-rendezvous v0.0.0-20200823014737-9f7001d12a5f // indirect
	github.com/yuin/gopher-lua v1.1.1 // indirect
	golang.org/x/sync v0.16.0 // indirect
)
