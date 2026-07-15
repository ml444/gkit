module github.com/ml444/gkit/cmd/protoc-gen-go-gorm/tests/user

go 1.23.4

replace (
	github.com/ml444/gkit => ../../../../
	github.com/ml444/gkit/cmd/protoc-gen-go-gorm => ../../
)

require (
	github.com/ml444/gkit v0.0.0-00010101000000-000000000000
	github.com/ml444/gkit/cmd/protoc-gen-go-gorm v0.0.0-00010101000000-000000000000
	google.golang.org/protobuf v1.36.8
	gorm.io/gorm v1.31.1
	gorm.io/hints v1.1.2
)

require (
	github.com/golang/protobuf v1.5.3 // indirect
	github.com/jinzhu/inflection v1.0.0 // indirect
	github.com/jinzhu/now v1.1.5 // indirect
	golang.org/x/sys v0.35.0 // indirect
	golang.org/x/text v0.28.0 // indirect
	google.golang.org/genproto/googleapis/rpc v0.0.0-20240108191215-35c7eff3a6b1 // indirect
	google.golang.org/grpc v1.60.1 // indirect
)
