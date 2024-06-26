module tests

go 1.21

replace github.com/ml444/gkit/cmd/protoc-gen-go-gorm => ../../protoc-gen-go-gorm

require (
	github.com/ml444/gkit/cmd/protoc-gen-go-gorm v0.0.0-00010101000000-000000000000
	google.golang.org/protobuf v1.32.0
	gorm.io/driver/mysql v1.5.2
	gorm.io/gorm v1.25.5
	gorm.io/hints v1.1.2
)

require (
	github.com/go-sql-driver/mysql v1.7.0 // indirect
	github.com/jinzhu/inflection v1.0.0 // indirect
	github.com/jinzhu/now v1.1.5 // indirect
	github.com/ml444/gkit v0.0.0-20240515210254-0a130fb39c1a // indirect
)
