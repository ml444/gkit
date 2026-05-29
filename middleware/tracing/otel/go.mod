module github.com/ml444/gkit/middleware/tracing/otel

go 1.23

replace (
	github.com/ml444/gkit => ../../..
	github.com/ml444/gkit/pkg/tracing => ../../../pkg/tracing
)

require (
	github.com/ml444/gkit v0.0.0-00010101000000-000000000000
	github.com/ml444/gkit/pkg/tracing v0.0.0-00010101000000-000000000000
	go.opentelemetry.io/otel v1.20.0
	go.opentelemetry.io/otel/trace v1.20.0
	google.golang.org/grpc v1.60.1
)

require (
	github.com/go-logr/logr v1.3.0 // indirect
	github.com/go-logr/stdr v1.2.2 // indirect
	github.com/go-playground/form v3.1.4+incompatible // indirect
	github.com/golang/protobuf v1.5.3 // indirect
	github.com/gorilla/mux v1.8.1 // indirect
	github.com/jinzhu/inflection v1.0.0 // indirect
	github.com/jinzhu/now v1.1.5 // indirect
	github.com/petermattis/goid v0.0.0-20230904192822-1876fd5063bc // indirect
	go.opentelemetry.io/otel/exporters/jaeger v1.17.0 // indirect
	go.opentelemetry.io/otel/metric v1.20.0 // indirect
	go.opentelemetry.io/otel/sdk v1.20.0 // indirect
	golang.org/x/net v0.32.0 // indirect
	golang.org/x/sync v0.10.0 // indirect
	golang.org/x/sys v0.28.0 // indirect
	golang.org/x/text v0.21.0 // indirect
	google.golang.org/genproto/googleapis/rpc v0.0.0-20240108191215-35c7eff3a6b1 // indirect
	google.golang.org/protobuf v1.34.2 // indirect
	gorm.io/gorm v1.25.7-0.20240204074919-46816ad31dde // indirect
)
