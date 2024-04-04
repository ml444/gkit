module tests

go 1.20

replace (
	github.com/ml444/gkit => ../../..
	github.com/ml444/gkit/cmd/protoc-gen-go-http => ../../protoc-gen-go-http
)

require (
	github.com/ml444/gkit v0.0.0-00010101000000-000000000000
	github.com/ml444/gkit/cmd/protoc-gen-go-http v0.0.0-00010101000000-000000000000
	github.com/xuri/excelize/v2 v2.8.1
	google.golang.org/genproto/googleapis/api v0.0.0-20231212172506-995d672761c0
	google.golang.org/protobuf v1.32.0
)

require (
	github.com/go-playground/form v3.1.4+incompatible // indirect
	github.com/golang/protobuf v1.5.3 // indirect
	github.com/google/go-cmp v0.6.0 // indirect
	github.com/gorilla/mux v1.8.1 // indirect
	github.com/jinzhu/inflection v1.0.0 // indirect
	github.com/jinzhu/now v1.1.5 // indirect
	github.com/ml444/gutil v0.0.0-20231221121703-d05adbb24fad // indirect
	github.com/mohae/deepcopy v0.0.0-20170929034955-c48cc78d4826 // indirect
	github.com/richardlehane/mscfb v1.0.4 // indirect
	github.com/richardlehane/msoleps v1.0.3 // indirect
	github.com/xuri/efp v0.0.0-20231025114914-d1ff6096ae53 // indirect
	github.com/xuri/nfp v0.0.0-20230919160717-d98342af3f05 // indirect
	golang.org/x/crypto v0.19.0 // indirect
	golang.org/x/net v0.21.0 // indirect
	golang.org/x/sys v0.17.0 // indirect
	golang.org/x/text v0.14.0 // indirect
	google.golang.org/genproto v0.0.0-20240102182953-50ed04b92917 // indirect
	google.golang.org/genproto/googleapis/rpc v0.0.0-20240108191215-35c7eff3a6b1 // indirect
	google.golang.org/grpc v1.60.1 // indirect
	gorm.io/gorm v1.25.5 // indirect
)
