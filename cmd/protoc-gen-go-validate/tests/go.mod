module github.com/ml444/gkit/cmd/protoc-gen-go-validate/tests

go 1.19

replace (
	github.com/ml444/gkit => ../../..
	github.com/ml444/gkit/cmd/protoc-gen-go-validate => ../
)

require (
	github.com/ml444/gkit v0.0.0-00010101000000-000000000000
	github.com/ml444/gkit/cmd/protoc-gen-go-validate v0.0.0-00010101000000-000000000000
	google.golang.org/protobuf v1.32.0
)

require (
	github.com/golang/protobuf v1.5.3 // indirect
	golang.org/x/sys v0.13.0 // indirect
	google.golang.org/genproto/googleapis/rpc v0.0.0-20240108191215-35c7eff3a6b1 // indirect
	google.golang.org/grpc v1.60.1 // indirect
)
