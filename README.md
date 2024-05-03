# gkit
[![Build Status](https://travis-ci.org/go-gadgets/gkit.svg?branch=master)](https://travis-ci.org/go-gadgets/gkit)
[![Coverage Status](https://coveralls.io/repos/github/go-gadgets/gkit/badge.svg?branch=master)](https://coveralls.io/github/go-gadgets/gkit?branch=master)
[![Go Report Card](https://goreportcard.com/badge/github.com/go-gadgets/gkit)](https://goreportcard.com/report/github.com/go-gadgets/gkit)
[![GoDoc](https://godoc.org/github.com/go-gadgets/gkit?status.svg)](https://godoc.org/github.com/go-gadgets/gkit)

[中文](README_CN.md)

`GKit` is a microservice framework that integrates `HTTP` and `GRPC` communication protocols. 
It aims at ease of use and encapsulates many daily WEB development components and tools to 
improve the development experience.
With the [gctl](https://github.com/ml444/gctl) tool, the basic code of the project can be 
quickly generated, so developers can focus more on business logic development.

## ProtocPlugin

The purpose of gkit is to use protobuf as much as possible to define and design API, 
and to achieve automatic generation of most codes through the following protoc plug-ins:
- `protoc-gen-go-errcode` Generate error codes and error messages, automatically obtain error messages and http status codes based on the error codes, and support multi-language configuration.
- `protoc-gen-go-field` Generate message structure field name constants and DB column names. You can use prefixes, suffixes or regular expressions to filter messages that need to generate field constants.
- `protoc-gen-go-gorm` Generate gorm model. The gorm tag of the field is defined through proto extension, as well as the serialization and deserialization methods of json and bytes type fields.
- `protoc-gen-go-http` Generate http routes and handlers, and use the `pluck.proto` extension to implement http request header settings, which is more useful when uploading and downloading files.
- `protoc-gen-go-validate` Generate parameter verification method to verify parameters according to the rules defined by `v.proto`. Call the middleware `validation.Validator` to start validation.

**Install protoc plugin**
```shell
$ go get -u github.com/ml444/gkit/cmd/protoc-gen-go-errcode \
    github.com/ml444/gkit/cmd/protoc-gen-go-field \
    github.com/ml444/gkit/cmd/protoc-gen-go-gorm \
    github.com/ml444/gkit/cmd/protoc-gen-go-http \
    github.com/ml444/gkit/cmd/protoc-gen-go-validate
```

## <a name="complete_example"></a>Usage example of protoc plug-in

```protobuf
syntax = "proto3";

package user;

option go_package = "pkg/user";

import "gkit/v/v.proto";
import "gkit/err/err.proto";
import "gkit/orm/orm.proto";
import "gkit/pluck/pluck.proto";
import "gkit/dbx/paging/paginate.proto";
import "google/api/annotations.proto";


service user {
    rpc CreateUser (CreateUserReq) returns (CreateUserRsp){
        option (google.api.http) = {
            post: "/v1/user"
            body: "*"
        };
    };
    rpc UpdateUser (UpdateUserReq) returns (UpdateUserRsp){
        option (google.api.http) = {
            put: "/v1/user"
            body: "*"
        };
    };
    rpc DeleteUser (DeleteUserReq) returns (DeleteUserRsp){
        option (google.api.http) = {
            delete: "/v1/user/{id}"
        };
    };
    rpc GetUser (GetUserReq) returns (GetUserRsp){
        option (google.api.http) = {
            get: "/v1/user/{id}"
        };
    };
    rpc ListUser (ListUserReq) returns (ListUserRsp){
        option (google.api.http) = {
            get: "/v1/user"
        };
    };

    rpc Upload (UploadReq) returns (UploadRsp){
        option (google.api.http) = {
            post: "/v1/storage/upload"
            body: "file_data"
        };
        option (pluck.request) = {
            default_headers: {      // Default request header settings
                content_type: "application/octet-stream"
            }
            headers_to: "file_info" // Extract the request header into the file info structure
        };
    };

    rpc Download (DownloadReq) returns (DownloadRsp){
        option (google.api.http) = {
            post: "/v1/storage/download"
            body: "*"
        };
        option (pluck.response) = {
            default_headers: {      // Default response header settings
                content_type: "application/vnd.openxmlformats"
                access_control_expose_headers: "Content-Disposition"
            }
            headers_from: "headers" // Set the fields in the `headers` map (or structure) to the http response header
            body_from: "data"       // Set the `data` field into the http response body
        };
    };
}

// range of error codes: [102000, 102999]
enum ErrCode {
    option (err.lower_bound) = 102000;
    option (err.upper_bound) = 102999;

    Success = 0;
    ErrIllegalParam = 102000  [(err.detail) = {status:400, message:"非法参数", polyglot: ["zh=非法参数", "en=Illegal parameters"]}];
    ErrParamRequired = 102001 [(err.detail) = {status:400, message:"缺失参数", polyglot: ["zh=缺失参数", "en=Missing parameters"]}];
    ErrNotFoundUser = 102002  [(err.detail) = {status:404, message:"未找到用户", polyglot: ["zh=未找到用户", "en=Record not found"]}];
}


message ModelUser {
    option (orm.enable) = true;
    option (orm.table_name) = "user";

    uint64 id = 1           [(orm.tags) = {comment: "主键", primary_key: true}];
    uint32 created_at = 101 [(orm.tags) = {comment: "创建时间"}];
    uint32 updated_at = 102 [(orm.tags) = {comment: "更新时间"}];
    uint32 deleted_at = 103 [(orm.tags) = {comment: "删除时间"}];
    string nick_name = 2    [(orm.tags) = {type: "varchar(50)", indexs: ["idx_age_name,priority:2"], comment: "昵称"}];
    string real_name = 3    [(orm.tags) = {type: "varchar(80)", comment: "真实姓名"}];
    string phone = 4        [(orm.tags) = {type: "varchar(25)", unique_indexs: "uidx_phone", not_null: true, comment: "名称"}];
    uint32 age = 5          [(orm.tags) = {type: "int", indexs: ["idx_age_name,priority:1"], comment: "年龄"}];
    uint32 sex = 6          [(orm.tags) = {type: "tinyint", comment: "性别"}];
    string email = 9        [(orm.tags) = {type: "varchar(255)", unique_indexs: "uidx_email", comment: "邮箱"}];
    string avatar = 10      [(orm.tags) = {type: "varchar(255)", comment: "头像"}];
}

message CreateUserReq {
    ModelUser data = 1 [(v.rules).message.required = true];     // Verify that the structure must pass a value
}
message CreateUserRsp {
    ModelUser data = 1;
}

message UpdateUserReq {
    uint64 id = 1       [(v.rules).uint64.gt = 0];              // Verify ID value must be greater than 0
    ModelUser data = 2  [(v.rules).message.required = true];    // Verify ModelUser must pass a value
}
message UpdateUserRsp {
    ModelUser data = 1;
}

message DeleteUserReq {
    uint64 id = 1                   [(v.rules).uint64.gte = 1];             // Verify ID value must be greater than 0
    // repeated uint64 id_list = 2  [(v.rules).repeated.min_items = 1];     // Verify that at least one ID is passed in
}
message DeleteUserRsp {}

message GetUserReq {
    uint64 id = 1       [(v.rules).uint64.gte = 1];          // Verify ID value must be greater than or equal to 1
}
message GetUserRsp {
    ModelUser data = 1;
}

message ListUserReq {
    repeated uint64 id_list = 1 [(v.rules).repeated.unique = true];                 // Check that the elements inside the array cannot be repeated
    optional string name = 2    [(v.rules).string = {min_len: 1, max_len: 50}];     // Verify that the string length is greater than or equal to 1 and less than or equal to 50
    optional string phone = 3   [(v.rules).string = {pattern: "\\d+", min_len:6, max_len: 25}];     // Verify that the string length is greater than or equal to 6, less than or equal to 25, and conforms to the regular expression
    optional string email = 4   [(v.rules).string.email = true];        // Verify whether it is the format of the email
    paging.Paginate paginate = 5;
}
message ListUserRsp {
    paging.Paginate paginate = 1;
    repeated ModelUser list = 2;
}

message UploadReq {
    message FileInfo {
        string file_name = 1;
        string file_suffix = 2;
    }
    FileInfo file_info = 1;
    bytes file_data = 2;
}
message UploadRsp {
    string url = 1;
    uint32 size = 2;
}


message DownloadReq {
    string filename = 1;
}
message DownloadRsp {
    map<string, string> headers = 1;
    bytes data = 2;
}
```

```shell
$ protoc --go_out=. \
       --go-grpc_out=. \
       --go-http_out=. \
       --go-field_out=. \
       --go-errcode_out=. \
       --go-validate_out=. \
        --proto_path=/your/path/gctl-templates/protos \
       user.proto
$ tree
.
├── user.pb.go
├── user.proto
├── user_errcode.pb.go
├── user_field.pb.go
├── user_grpc.pb.go
├── user_http.pb.go
├── user_orm.pb.go
└── user_validate.pb.go
```

```
import "gkit/v/v.proto";
import "gkit/err/err.proto";
import "gkit/orm/orm.proto";
import "gkit/pluck/pluck.proto";
import "gkit/dbx/paging/paginate.proto";
```

The internal import of proto is referenced`gctl-templates/protos/gkit`,
If you put these import files elsewhere, you can modify it to `import "your/path/xxx.proto`

## Project structure introduction

```
.
├── cmd
│   ├── protoc-gen-go-errcode
│   ├── protoc-gen-go-field
│   ├── protoc-gen-go-gorm
│   ├── protoc-gen-go-http
│   └── protoc-gen-go-validate
├── config
├── dbx
├── errorx
├── log
├── middleware
│   ├── general
│   ├── logging
│   ├── ratelimit
│   ├── recovery
│   ├── trace
│   └── validate
├── optx     
├── pkg
│   ├── auth
│   ├── env
│   ├── header
│   ├── routine
│   └── tracing
├── transport
├── go.mod
└── go.sum
```
- **cmd**: protoc plugins
- **config**: The configuration module defines the reading method of configuration items through structure tags, and supports configuration information obtained from the command line, environment variables, yaml, json, toml, etc.
- **errorx**: The error handling module encapsulates the error handling methods in daily development, supports custom error codes and error messages, and supports automatically obtaining error messages and http status codes based on error codes.
- **dbx**: The database module is secondary encapsulated based on gorm, encapsulates the chain method of query (Eq\In\Like...), and supports soft deletion and paging query.
- **optx**: It defines the conditional filtering method of list data, encapsulates its processing method for the two parameter passing methods (enumeration and pointer) of list query, and encapsulates its processor module.
- **log**: The log module defines a log interface and outputs to standard output by default. The log implementation can be customized.Also encapsulates gorm’s log output,unified output to the specified logger.
- **middleware**: The middleware module, It mainly includes middleware such as request and response logs, ratelimit, recovery, tracking, and parameter verification.
- **transport**: Communication transport module,It mainly includes the transport modules of http and grpc.
- **pkg**: The public module includes some basic tool classes, such as authentication, environment judgment, request headers, coroutine security processing, link tracking, etc.

## Core module description
### errorx

The error handling module mainly includes the following functions:

- Register error codes and error messages.
- Encapsulate error codes, error messages, error details, and error stacks.
- Automatically obtain error information and http status codes based on custom error codes.
- You can set error messages in multiple languages and return the corresponding error messages based on the `Accept-Language` in the request header.
- Determine whether it is the specified error code based on the custom error code.
- Convert error code to http status code according to GRPC.
- Convert HTTP status code to GRPC error code.

You can define the error code and error information in the proto file, 
and then register the error code and error information through `errorx.Register Err Code()`.
Then instantiate the Error object through methods such as errorx.New(). 
When the request returns an error, the corresponding http status code and 
error information will be returned.

```protobuf
syntax = "proto3";

package user;

import "gkit/err/err.proto";     // Source File: github.com/ml444/gkit/cmd/protoc-gen-go-errcode/err/err.proto

// range of error codes: [102000, 102999]
enum ErrCode {
    option (err.lower_bound) = 101000;
    option (err.upper_bound) = 101999;

    Success = 0;
    ErrIllegalParam = 102000 [(err.detail) = {status:400, message:"Illegal parameters"}];
    ErrParamRequired = 102001 [(err.detail) = {status:400, message:"Missing parameters"}];
    ErrNotFoundUser = 102002 [(err.detail) = {status:404, message:"The user not found"}];
    // or
    // ErrIllegalParam = 102000 [(err.detail) = {status:400, polyglot: ["zh=非法参数", "en=Illegal parameters"]}];
    // ErrParamRequired = 102001 [(err.detail) = {status:400, polyglot: ["zh=缺失参数", "en=Missing parameters"]}];
    // ErrNotFoundUser = 102002 [(err.detail) = {status:404, polyglot: ["zh=未找到用户", "en=The user not found"]}];
}
```

```shell
$ protoc --go_out=. --go-errcode_out=. user.proto
$ tree
.
├── user.pb.go
├── user.proto
└── user_errcode.pb.go
```

```go
package main

import (
	"context"
	"github.com/ml444/gkit/errorx"
	"gitlab.xxx.com/mygroup/myproject/pkg/user"
)

type UserService struct {
	user.UnsafeUserServer
}

func NewUserService() UserService {
	return UserService{}
}
func (s *UserService) GetUser(ctx context.Context, req *user.GetUserReq) (*user.GetUserRsp, error) {
	if req.Id == 0 {
		return nil, errorx.New(user.ErrParamRequired) // response: {"status_code": 400, "error_code": 102001, "message": "缺失参数"}
	}
	// do something

	// if not found user
	return nil, errorx.New(user.ErrNotFoundUser) // response: {"status_code": 404, "error_code": 102003, "message": "未找到用户"}
}

func main() {
	errorx.RegisterError(user.ErrCodeMap)
	// pass
}
```

### dbx

Secondary encapsulation based on gorm mainly includes the following functions:

- Encapsulates the creation, update, query, and deletion of gorm. The query encapsulates the chain method (Eq\Gt\Lt\In\Not In\Between...), making it easier to use, and supports soft deletion and paging query.
- The parameter structure `QueryOpts` that encapsulates complex queries can make it easier to process query conditions under some complex queries.
- For `Not Found Record` error handling, the error code and error message can be customized.
- The `dbx.paging` module of list paging query makes paging list query easier to use.

Basic usage：

```go
package main

import (
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"github.com/ml444/gkit/dbx"
	"github.com/ml444/gkit/dbx/paging"
)

type ModelUser struct {
  Id        uint64 `gorm:"comment:主键;primarykey"`
  CreatedAt uint32 `gorm:"comment:创建时间"`
  UpdatedAt uint32 `gorm:"comment:更新时间"`
  DeletedAt uint32 `gorm:"comment:删除时间"`
  Name      string `gorm:"type:varchar(50);comment:名称;index:idx_age_name,priority:2"`
  Phone     string `gorm:"not null;type:varchar(25);comment:名称;uniqueIndex:uidx_phone"`
  Age       uint32 `gorm:"type:int;comment:年龄;index:idx_age_name,priority:1"`
  Sex       uint32 `gorm:"type:tinyint;comment:性别"`
  Email     string `gorm:"type:varchar(255);comment:邮箱;uniqueIndex:uidx_email"`
  Avatar    string `gorm:"type:varchar(255);comment:头像"`
}

type GroupBy struct {
	Name  string `json:"name"`
	Total uint32 `json:"total"`
}

func getDB() *gorm.DB {
	db, err := gorm.Open(mysql.Open("user:password@tcp(192.168.1.100:3306)/gkit?charset=utf8&parseTime=True&loc=Local"), &gorm.Config{})
	if err != nil {
		panic(err)
	}
	return db
}

func main() {
	var err error
	scope := dbx.NewScope(getDB(), &ModelUser{})
	// create data: INSERT INTO `model_user` (`name`,`age`,`created_at`,`updated_at`) VALUES ('test',18,1625673600,1625673600)
	err = scope.Create(&ModelUser{Name: "test", Age: 18})

	// update data: UPDATE `model_user` SET `name`='test2',`age`=20,`updated_at`=1625673600 WHERE `id` = 1
	err = scope.Eq("id", 1).Update(&ModelUser{Name: "test2", Age: 20})

	// soft delete data: UPDATE `model_user` SET `deleted_at`=1625673600 WHERE `id` IN (1,2,3)
	err = scope.In("id", []uint64{1, 2, 3}).Delete()

	// select data: SELECT * FROM `model_user` WHERE `id` = 1 AND `deleted_at` = 0 LIMIT 1
	var user ModelUser
	err = scope.Eq("id", 1).First(&user)
	err = scope.SetNotFoundErr(notFoundErrCode).Eq("id", 1).First(&user)
	// if not found record, return error: errorx.New(notFoundErrCode)

	// select data: SELECT * FROM `model_user` WHERE `deleted_at` = 0 AND `name` Like 'test%' AND `age` <= 25 LIMIT 10 OFFSET 0
	var users []*ModelUser
	err = scope.LikePrefix("name", "test").Lte("age", 25).Limit(10).Offset(0).Find(&users)
	// Or use paginate query to get total count
	paginate, err := scope.LikePrefix("name", "test").Lte("age", 25).PaginateQuery(&paging.Paginate{Page: 1, Size: 10, SkipCount: false}, &users)
	// paginate: Paginate{Total: 100, Page: 1, Size: 10} 

	// GroupBy and Having
	var userGroup []*GroupBy
	err = scope.Select("name", "count(*) AS total").Group("name").Having("age > 18").Find(&userGroup)

	// OrderBy
	err = scope.Order("age DESC").Find(&users)
}
```

#### Paging query

The data structure of paging query is defined through `paging/paginate.proto`, 
and different paging methods can be selected according to the actual situation.
There are also two ways to use pagination queries:

1. Query by specifying the number of pages and the number of pages per page.
2. Scroll page query, this method is suitable for queries with large amounts of data.

_**Note**_: During paging query, you can call the `skip count` parameter after 
the second page to save database performance. Of course, this requires the 
front-end engineer to cache the total number obtained for the first page.

Paginated database queries can use the `Scope.PaginateQuery()` method, 
which internally calls `Count()` and `Find()`.

**Proto definition of paging mode**：

```protobuf
syntax = "proto3";
import "dbx/paging/paging.proto";
/*
message Paginate {
  uint32 page = 1;
  uint32 size = 2;
  int64 total = 3;
  bool skip_count = 4;
}
 */

message ListUserReq {
    paging.Paginate paginate = 1;   // Number of pages and quantity per page must be required 
}

message ListUserRsp {
    paging.Paginate paginate = 1;
}
```

**proto definition of scrolling mode**：

```protobuf
syntax = "proto3";
import "dbx/paging/paging.proto";
message ListUserReq {
    paging.Scroll scroll = 1;     // Scroll page query
}
message ListUserRsp {
    repeated ModelUser list = 2;
}
```

### optx

Filter filter query module. There are two ways (enumeration and pointer) to 
define the query parameters of the list. And they all encapsulate the 
corresponding processor `Processor` and the corresponding method of 
processing parameters, and standardize the processing of list query parameters.


This module encapsulates two list parameter filtering query methods, 
one is to define query parameters through `optx.Options`, 
and the other is to directly define pointer parameters to query.
Which method to use can be chosen according to the actual situation. 
If it is an API that hides query parameters from the outside, 
you can use `optx.Options` to define query parameters, which can better 
control the query parameters and conceal the meaning of the parameters. 
because its parameters are represented by enumeration values.
In general situations or scenarios where zero values need to be passed, 
you can directly define pointer parameters to query. 
This method is straightforward.

**Define query parameters in enumeration mode**：
~~TODO: I think no one will like this method. If someone likes it, I will consider implementing a proto plug-in to generate this kind of code.~~

```protobuf
syntax = "proto3";
import "optx/optx.proto";
message ListUserReq {
    enum ListOpt {
        ListOptNil = 0;
        // @valueType: uint64List
        ListOptIdList = 1;
        // @valueType: string
        ListOptLikeName = 2;
        // @valueType: string
        ListOptPhone = 4;
    }
    // @ref_to: ListUserReq.ListOpt
    optx.Options list_option = 1;
}
message ListUserRsp {
    repeated ModelUser list = 2;
}
```

```go
package main

import (
	"context"
	"github.com/ml444/gkit/optx"
	"gitlab.xxx.com/mygroup/myproject/pkg/user"
)

type UserService struct {
	user.UnsafeUserServer
}

func NewUserService() UserService {
	return UserService{}
}

func (s *UserService) ListUser(ctx context.Context, req *user.ListUserReq) (*user.ListUserRsp, error) {
	var users []*ModelUser
	scope := getDBScope()
	err := optx.NewProcessor(req.ListOption).
		AddUint64List(user.ListUserReq_ListOptIdList, func(valList []uint64) error {
			scope.In("id", ids)
			return nil
		}).
		AddString(user.ListUserReq_ListOptLikeName, func(val string) error {
			scope.Like("name", val)
			return nil
		}).
		AddString(user.ListUserReq_ListOptPhone, func(phone string) error {
			scope.Eq("phone", phone)
			return nil
		}).
		Process()
	if err != nil {
		return nil, err
	}
	err = scope.Find(&users)
	if err != nil {
		return nil, err
	}
	// do something
	return &user.ListUserRsp{List: users}, nil
}
```

**Define query parameters in pointer mode**：

```protobuf
message ListUserReq {
    repeated uint64 id_list = 1;
    optional string like_name = 2;
    optional string phone = 3;
}
message ListUserRsp {
    repeated ModelUser list = 2;
}
```

```go
package main

import (
	"context"
	"github.com/ml444/gkit/optx"
	"gitlab.xxx.com/mygroup/myproject/pkg/user"
)

type UserService struct {
	user.UnsafeUserServer
}

func NewUserService() UserService {
	return UserService{}
}

func (s *UserService) ListUser(ctx context.Context, req *user.ListUserReq) (*user.ListUserRsp, error) {
	var users []*user.ModelUser
	scope := dbUser.Scope()
	err := optx.NewPtrProcessor().
		AddHandle(user.FieldIdList, func(val interface{}) error {
			scope.In(user.DbFieldId, val)
			return nil
		}).
		AddHandle(user.FieldLikeName, func(val interface{}) error {
			scope.Like(user.DbFieldName, val.(string))
			return nil
		}).
		AddHandle(user.FieldPhone, func(val interface{}) error {
			scope.Eq(user.DbFieldPhone, val)
			return nil
		}).
		Process(req)
	if err != nil {
		return nil, err
	}
	err = db.Find(&users)
	if err != nil {
		return nil, err
	}
	return &user.ListUserRsp{List: users}, nil
}
```


[List User example of proto](#complete_example),contains database query, filtering and paging functions:

```go
package main

import (
	"context"

	"github.com/ml444/gkit/dbx"
	"github.com/ml444/gkit/log"
	"github.com/ml444/gkit/optx"

	"gitlab.xxx.com/xxx/internal/db"
	"gitlab.xxx.com/xxx/pkg/user"
)

type UserService struct {
	user.UnsafeUserServer
}

func NewUserService() UserService {
	return UserService{}
}

/*
...Omit other code...
*/

func (s UserService) ListUser(ctx context.Context, req *user.ListUserReq) (*user.ListUserRsp, error) {
	var rsp user.ListUserRsp

	scope := dbx.NewScope(db.DB(), &user.ModelUser{})
	err := optx.NewPtrProcessor().
		AddHandle(user.FieldIdList, func(val interface{}) error {
			scope.In(user.DbFieldId, val)
			return nil
		}).
		AddHandle(user.FieldName, func(val interface{}) error {
			scope.Like(user.DbFieldNickName, val.(string))
			return nil
		}).
		AddHandle(user.FieldPhone, func(val interface{}) error {
			scope.Eq(user.DbFieldPhone, val)
			return nil
		}).
		Process(req)

	// do something...
	rsp.Paginate, err = scope.PaginateQuery(req.Paginate, &rsp.List)
	if err != nil {
		log.Errorf("err: %v", err)
		return nil, err
	}
	return &rsp, nil
}
```

### log

The log module mainly includes the following functions:

- The log interface is defined and the log implementation can be customized. By default, `os.stdout` is used for output.
- Customize the log output of gorm, combine it with the custom log implementation, and output the specified location together.

Use directly：

```go
package main

import (
	"github.com/ml444/gkit/log"
)

func main() {
	log.Debug("this is debug")
	log.Info("this is info")
	log.Warn("this is warn")
	log.Error("this is error")
	log.Fatal("this is error")

	name := "foo"
	log.Debugf("hi, %s! this is debug", name)
	log.Infof("hi, %s! this is info", name)
	log.Warnf("hi, %s! this is warn", name)
	log.Errorf("hi, %s! this is error", name)
	log.Fatalf("hi, %s! this is fatal", name)
}
```

Example using [glog](https://github.com/ml444/glog)

```go
package main

import (
	"github.com/ml444/gkit/log"
	glog "github.com/ml444/glog"
	glogconf "github.com/ml444/glog/config"
	gloglevel "github.com/ml444/glog/level"
)

func InitLogger(debug bool) error {
	err := log.InitLog(
		logconf.SetLoggerName("myname"),
		logconf.SetLevel2Logger(level.InfoLevel),
		logconf.SetFileDir2Logger("./log"),
		func(config *logconf.Config) {
			config.Handler.LogHandlerConfig.Formatter.Text.DisableColors = !debug
		},
	)
	if err != nil {
		return err
	}
	if debug {
		err = log.InitLog(logconf.SetLevel2Logger(level.DebugLevel))
		if err != nil {
			return err
		}
	}
	log.SetLogger(glog.GetLogger())
	return nil
}

func main() {
	err := InitLogger(true)
	if err != nil {
		println(err.Error())
		return
	}

	log.Debug("this is debug")
	log.Info("this is info")
	log.Warn("this is warn")
	log.Error("this is error")
	log.Fatal("this is error")

	name := "foo"
	log.Debugf("hi, %s! this is debug", name)
	log.Infof("hi, %s! this is info", name)
	log.Warnf("hi, %s! this is warn", name)
	log.Errorf("hi, %s! this is error", name)
	log.Fatalf("hi, %s! this is fatal", name)
}
```

### middleware

Middleware, the following middleware is currently developed:

- `general`: used to handle empty responses and unified error output in responses.
- `logging` : record request logs, response logs, request and response logs.
- `ratelimit` : used to limit the access frequency of requests.
- `recovery` : used to recover panic and record logs.
- `tracking` : used for tracking request links.
- `validation`: used for parameter verification. Only when this middleware is enabled, the parameter verification rules defined in the proto file will take effect.

```go
package main

import (
	"github.com/ml444/gkit/middleware/general"
	"github.com/ml444/gkit/middleware/logging"
	"github.com/ml444/gkit/middleware/ratelimit"
	"github.com/ml444/gkit/middleware/recovery"
	"github.com/ml444/gkit/middleware/trace"
	"github.com/ml444/gkit/middleware/validate"
	"github.com/ml444/gkit/transport/httpx"
)

func main() {
	// HTTP
	httpx.NewServer(
		httpx.Address(":5050"),
		httpx.Middleware(
			//  Handling empty responses
			general.ReplaceEmptyResponse(struct {
				StatusCode int32
				ErrCode    int32
				Message    string
			}{200, 0, "success"}),
			// Unify errors into the errorx.Error structure
			general.WrapError(),
			// Record the input parameters and time consumption of the request
			logging.LogRequest(),
			// Frequency limiting middleware 
			ratelimit.FrequencyLimit(
				&ratelimit.LimitCfg{
					Kind: ratelimit.MatchKindAll,
					Freqs: []*ratelimit.Cycle{
						{Period: time.Second * 1, Limit: 100},
					},
				},
				&ratelimit.LimitCfg{
					Paths: []string{user.OperationUserGetUser},
					Freqs: []*ratelimit.Cycle{
						{Period: time.Second * 1, Limit: 50},
						{Period: time.Second * 60, Limit: 3000},
					},
				},
			),
			// Recovery middleware
			recovery.Recovery(),
			// Tracking middleware
			trace.Server(),
			// Verification middleware
			validate.Validator(),

		),
	)
}
```

### transport

- Convert the core logic of the http request into the service method of grpc.
- It encapsulates the middleware of http and grpc and unifies the middleware interface.

### pkg

Public modules include some basic tools or functions


