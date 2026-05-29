module github.com/ml444/gkit/middleware/ratelimit/redis

go 1.23

replace github.com/ml444/gkit => ../../..

require (
	github.com/alicebob/miniredis/v2 v2.38.0
	github.com/ml444/gkit v0.0.0-00010101000000-000000000000
	github.com/redis/go-redis/v9 v9.14.0
)

require (
	github.com/cespare/xxhash/v2 v2.3.0 // indirect
	github.com/dgryski/go-rendezvous v0.0.0-20200823014737-9f7001d12a5f // indirect
	github.com/golang/protobuf v1.5.3 // indirect
	github.com/yuin/gopher-lua v1.1.1 // indirect
	golang.org/x/sys v0.28.0 // indirect
	google.golang.org/genproto/googleapis/rpc v0.0.0-20240108191215-35c7eff3a6b1 // indirect
	google.golang.org/grpc v1.60.1 // indirect
	google.golang.org/protobuf v1.34.2 // indirect
)
