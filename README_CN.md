# gkit

gkit是一个集成`HTTP`和`GRPC`通信协议的微服务框架，主旨在于易使用，通过封装日常WEB开发的组件，
配合`gctl`工具，可以快速的生成项目的基础代码，从而开发人员可以更专注于业务逻辑开发。

gkit的宗旨是尽可能的使用protobuf来定义和设计接口，通过以下的多个protoc插件来实现大部分代码的自动生成：

- `protoc-gen-go-errcode` 生成错误码和错误信息，可以根据错误码自动获取错误信息和http状态码，支持多语言配置。
- `protoc-gen-go-field` 生成字段定义，可用前缀、后缀或正则表达式来筛选需要生成字段常量的message。
- `protoc-gen-go-gorm` 生成gorm的model，包含了字段的tag定义，以及json和bytes类型字段的序列化和反序列化方法。
- `protoc-gen-go-http` 生成http的路由和handler，并且可用pluck模块实现http的请求头设置，这在上传和下载文件中比较有用。
- `protoc-gen-go-validate` 生成参数校验的方法，可以根据`v.proto`定义的规则来校验参数。

```shell
$ go get -u github.com/ml444/gkit/cmd/protoc-gen-go-errcode \
    github.com/ml444/gkit/cmd/protoc-gen-go-field \
    github.com/ml444/gkit/cmd/protoc-gen-go-gorm \
    github.com/ml444/gkit/cmd/protoc-gen-go-http \
    github.com/ml444/gkit/cmd/protoc-gen-go-validate
```

## <a name="complete_example"></a>完整示例

protoc插件的使用示例：

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
            default_headers: {
                content_type: "application/octet-stream"
            }
            headers_to: "file_info"
        };
    };

    rpc Download (DownloadReq) returns (DownloadRsp){
        option (google.api.http) = {
            post: "/v1/storage/download"
            body: "*"
        };
        option (pluck.response) = {
            default_headers: {
                content_type: "application/vnd.openxmlformats"
                access_control_expose_headers: "Content-Disposition"
            }
            headers_from: "headers"
            body_from: "data"
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
    ModelUser data = 1 [(v.rules).message.required = true];
}
message CreateUserRsp {
    ModelUser data = 1;
}

message UpdateUserReq {
    uint64 id = 1 [(v.rules).uint64.gt = 0];
    ModelUser data = 2 [(v.rules).message.required = true];
}
message UpdateUserRsp {
    ModelUser data = 1;
}

message DeleteUserReq {
    uint64 id = 1 [(v.rules).uint64.gte = 1];
    // repeated uint64 id_list = 2 [(v.rules).repeated.min_items = 1];
}
message DeleteUserRsp {}

message GetUserReq {
    uint64 id = 1 [(v.rules).uint64.gt = 0];
}
message GetUserRsp {
    ModelUser data = 1;
}

message ListUserReq {
    repeated uint64 id_list = 1 [(v.rules).repeated.unique = true];
    optional string name = 2    [(v.rules).string = {min_len: 1, max_len: 50}];
    optional string phone = 3   [(v.rules).string = {pattern: "\\d+", min_len:6, max_len: 25}];
    optional string email = 4   [(v.rules).string.email = true];
    paging.Paginate paginate = 5;
}
message ListUserRsp {
    paging.Paginate paginate = 1;
    repeated ModelUser list = 2;
}

message UploadReq {
    message FileInfo {
        // @desc: 文件名称
        string file_name = 1;
        // @desc: 文件后缀
        string file_suffix = 2;
    }
    FileInfo file_info = 1;
    // @desc: 上传的文件内容
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

## gkit项目结构介绍

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

- **cmd**: protoc插件
- **config**: 配置模块，通过结构体tag来定义配置项的读取方式，支持从命令行、环境变量、yaml、json、toml等方式来获取的配置信息。
- **errorx**: 错误处理模块，封装了日常开发中的错误处理方式，支持自定义错误码和错误信息，支持根据错误码自动获取错误信息和http状态码。
- **dbx**: 基于gorm进行二次封装的数据库模块，封装了增删改查的方法，支持软删除和分页查询。
- **optx**: 定义了列表数据的条件筛选方式，对列表查询的两种（枚举和指针）传参方式封装了其处理方法；并封装了其处理器模块。
- **log**: 日志模块，定义了日志接口，可以自定义日志实现。把gorm的日志输出也封装了起来，统一输出到指定的logger。
- **middleware**: 中间件模块, 主要包含了请求和响应日志、限流、恢复、跟踪、参数校验等中间件。
- **transport**: 通信传输模块，主要包含了http和grpc的传输模块。
- **pkg**: 公共模块，包含了一些基础的工具类，如：认证、环境判断、请求头、协程安全处理、链路追踪等。

## 核心模块说明

### errorx

错误处理模块，主要包含以下几个功能：

- 注册错误码和错误信息
- 封装错误码和错误信息和错误详情和错误堆栈
- 可根据自定义错误码自动获取错误信息和http状态码
- 根据自定义错误码判断是否为指定的错误码
- 根据GRPC的错误码转化为http状态码
- 根据http状态码转化为GRPC的错误码

- 可以通过在proto文件中定义错误码和错误信息，然后通过errorx.RegisterErrCode()注册错误码和错误信息，
  然后通过errorx.New()生成错误信息。在接口返回错误码时会主动转化为http状态码，和错误的中文信息。

```protobuf
syntax = "proto3";

package user;

import "gkit/err/err.proto";     // 引入错误码定义文件: github.com/ml444/gkit/cmd/protoc-gen-go-errcode/err/err.proto

// range of error codes: [102000, 102999]
enum ErrCode {
    option (err.lower_bound) = 101000;
    option (err.upper_bound) = 101999;

    Success = 0;
    ErrIllegalParam = 102000 [(err.detail) = {status:400, message:"非法参数"}];
    ErrParamRequired = 102001 [(err.detail) = {status:400, message:"缺失参数"}];
    ErrNotFoundUser = 102002 [(err.detail) = {status:404, message:"未找到用户"}];
    // or
    // ErrIllegalParam = 102000 [(err.detail) = {status:400, polyglot: ["zh=非法参数", "en=Illegal parameters"]}];
    // ErrParamRequired = 102001 [(err.detail) = {status:400, polyglot: ["zh=缺失参数", "en=Missing parameters"]}];
    // ErrNotFoundUser = 102002 [(err.detail) = {status:404, polyglot: ["zh=未找到用户", "en=Record not found"]}];
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

基于gorm的数据库模块和Proto格式的model做了一些预处理的工作，主要包含以下几个功能：

- 封装gorm的增删改查，使其更易于使用，并支持软删除以及分页查询。
- 针对NotFoundRecord的错误处理，可以自定义错误码和错误信息。
- 封装列表的分页查询`dbx.paging`，使其更易于使用。

基础使用：

```go
package main

import (
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"github.com/ml444/gkit/dbx"
	"github.com/ml444/gkit/dbx/paging"
)

type ModelUser struct {
	ID        uint64 `gorm:"primary_key"`
	Name      string
	Age       uint8
	CreatedAt uint32
	UpdatedAt uint32
	DeletedAt uint32
}

type GroupBy struct {
	Name  string `json:"name"`
	Total uint32 `json:"total"`
}

func getDB() *gorm.DB {
	db, err := gorm.Open(mysql.Open("dsn"), &gorm.Config{})
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

	// delete data: UPDATE `model_user` SET `deleted_at`=1625673600 WHERE `id` IN (1,2,3)
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
	// return paginate: Paginate{Total: 100, Page: 1, Size: 10} 

	// GroupBy and Having
	var userGroup []*GroupBy
	err = scope.Select("name", "count(*) AS total").Group("name").Having("age > 18").Find(&userGroup)

	// OrderBy
	err = scope.Order("age DESC").Find(&users)
}
```

[完整示例](#complete_example)的proto的示例：

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
...省略其他代码...
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

#### 分页查询

分页查询的参数和结果返回，都是通过`paging.Paginate`来定义的，可以根据实际情况来选择是否使用分页查询。
分页查询也有三种使用方式：

1. 通过指定页数和每页数量来查询，这种方式适合于前端分页查询。
2. 通过指定偏移量和每页数量来查询，这种方式适合于后端分页查询。
3. 滚动翻页查询，这种方式适合于大数据量的查询。

4. 分页查询也可以通过`Scope.PaginateQuery()`来实现，这样可以更方便使用。

**分页方式的proto定义**：

```protobuf
syntax = "proto3";
import "dbx/paging/paging.proto";
/*
message Paginate {
  // current page
  uint32 page = 1;
  // page size
  uint32 size = 2;
  // offset is the starting point of the table index.
  uint32 offset = 3;
  // total number of data
  int64 total = 5;
  // When SkipCount is true,
  // even if CurrentPage is equal to 1, don't count the total.
  bool skip_count = 6;
}
 */

message ListUserReq {
    paging.Paginate paginate = 1;   // 指定页数和每页数量 或 指定偏移量和每页数量
}

message ListUserRsp {
    paging.Paginate paginate = 1;
}
```

**滚动方式的proto定义**：

```protobuf
syntax = "proto3";
import "dbx/paging/paging.proto";
message ListUserReq {
    paging.Scroll scroll = 1;     // 滚动翻页查询
}
message ListUserRsp {
    repeated ModelUser list = 2;
}
```

### optx

列表查询模块，主要包含以下几个功能：

- 通过枚举和指针两种方式来定义列表的参数查询。
- 封装查询处理器`Processor`和封装处理参数的方法，规范化处理列表查询参数。

本模块封装了两种列表参数筛选查询方式，一种是通过`optx.Options`来定义查询参数，另一种是通过直接定义指针参数来查询。
具体使用哪种方式，可以根据实际情况来选择，如果是对外提供的接口，建议使用`optx.Options`来定义查询参数，这样可以更好的控制查询参数，并具备隐蔽性。
如果是内部接口或者需要传递零值的场景，可以直接定义指针参数来查询，这样可以更方便的使用。

**枚举方式定义查询参数**：
TODO: 我觉得这种方式应该不会有人喜欢，如果有人喜欢我会考虑实现一个proto插件来生成这种代码。

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

**指针方式定义查询参数**：

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

### log

日志模块，主要包含以下几个功能：

- 定义了日志接口，可以自定义日志实现
- 自定义gorm的日志输出

```go
package main

func main() {
	// 初始化日志
	log.Init()
	// 初始化日志并设置日志级别
	log.Init(log.WithLevel(log.LevelDebug))
	// 自定义gorm的日志输出
	log.Init(log.WithGormLogger(func(l *log.Logger) {
		l.SetFormatter(&logrus.JSONFormatter{})
		l.SetOutput(os.Stdout)
		l.SetLevel(logrus.DebugLevel)
	}))
}
```

### middleware

中间件，目前开发了以下中间件：
- 通用中间件：用于处理响应中的空响应和统一错误输出的的通用中间件
- 日志中间件： 记录请求日志、响应日志、请求和响应日志
- 限流中间件： 用于限制请求的访问频率
- 恢复中间件： 用于恢复panic，并记录日志
- 跟踪中间件： 用于链路追踪
- 参数校验中间件： 用于参数校验，只有启用这个中间件，在proto文件中定义的参数校验规则才会生效

```go
package main

import (
    "github.com/ml444/gkit/middleware/general"
    "github.com/ml444/gkit/middleware/logging"
    "github.com/ml444/gkit/middleware/ratelimit"
    "github.com/ml444/gkit/middleware/recovery"
    "github.com/ml444/gkit/middleware/trace"
    "github.com/ml444/gkit/middleware/validate"
)

func main() {
    // 通用中间件
    general.NewGeneralMiddleware()
    // 日志中间件
    logging.NewLoggingMiddleware()
    // 限流中间件
    ratelimit.NewRateLimitMiddleware()
    // 恢复中间件
    recovery.NewRecoveryMiddleware()
    // 跟踪中间件
    trace.NewTraceMiddleware()
    // 参数校验中间件
    validate.NewValidateMiddleware()
}
```
  

### transport

传输模块，主要包含以下几个功能：

- http传输
    - 把http请求转化为grpc请求
    - 根据传入的grpc Service生成http路由
    - http请求头的转换为context内容
    - 前置中间件
- grpc传输
  - xds负载均衡

### pkg

公共模块，包含了一些基础的工具类，主要包含以下几个模块：

- **auth**: 基于jwt的认证模块，主要包含以下几个功能：
    - 生成/解析/验证`access token`和`refresh token`
    - 设置过期时间
    - 自定义claim
- **env**: 用于工作环境的判断，如：是否为开发环境、测试环境、生产环境等
- **header**: 请求头相关的处理及httpHeader与Context的转换
- **routine**: 用于协程的安全处理，在协程中使用`routine.Go()`代替`go`关键字，可以捕获协程中的panic，并记录日志。
- **tracing**: 链路追踪模块，主要包含以下几个功能：
    - 生成链路追踪的`trace id`和`span id`
    - 从http请求头中获取链路追踪的`trace id`和`span id`
    - 设置链路追踪的`trace id`和`span id`到http请求头

