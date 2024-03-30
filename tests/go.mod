module github.com/ml444/gkit/tests

go 1.19

//go.etcd.io/bbolt => github.com/coreos/bbolt v1.3.5
replace github.com/ml444/gkit => ../

require (
	github.com/envoyproxy/protoc-gen-validate v1.0.2
	github.com/ml444/gkit v0.0.0-00010101000000-000000000000
	google.golang.org/genproto/googleapis/api v0.0.0-20231212172506-995d672761c0
	google.golang.org/grpc v1.60.1
	google.golang.org/protobuf v1.32.0
	gorm.io/driver/sqlite v1.5.4
	gorm.io/gorm v1.25.5
)

require (
	github.com/cncf/xds/go v0.0.0-20231128003011-0fa0005c9caa // indirect
	github.com/envoyproxy/go-control-plane v0.11.2-0.20230627204322-7d0032219fcb // indirect
	github.com/go-playground/form v3.1.4+incompatible // indirect
	github.com/golang/protobuf v1.5.3 // indirect
	github.com/gorilla/mux v1.8.1 // indirect
	github.com/jinzhu/inflection v1.0.0 // indirect
	github.com/jinzhu/now v1.1.5 // indirect
	github.com/mattn/go-sqlite3 v1.14.17 // indirect
	github.com/ml444/gutil v0.0.0-20231221121703-d05adbb24fad // indirect
	golang.org/x/net v0.17.0 // indirect
	golang.org/x/sys v0.13.0 // indirect
	golang.org/x/text v0.13.0 // indirect
	google.golang.org/genproto v0.0.0-20240102182953-50ed04b92917 // indirect
	google.golang.org/genproto/googleapis/rpc v0.0.0-20240108191215-35c7eff3a6b1 // indirect
)
