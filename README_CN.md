# gkit

[English](README.md)


gkit是一个集成`HTTP`和`GRPC`通信协议的微服务框架，易用性为目标，封装许多日常WEB开发的组件及工具以提高开发体验，
配合`gctl`工具，可以快速的生成项目的基础代码，从而开发人员可以更专注于业务逻辑开发。

## protoc插件

gkit的宗旨是尽可能的使用protobuf来定义和设计接口，通过以下的多个protoc插件来实现大部分代码的自动生成：

- [**protoc-gen-go-errcode**](https://github.com/ml444/gkit/tree/master/cmd/protoc-gen-go-errcode): 生成错误码和错误信息，可以根据错误码自动获取错误信息和http状态码，支持多语言配置。
- [**protoc-gen-go-field**](https://github.com/ml444/gkit/tree/master/cmd/protoc-gen-go-field): 生成message的字段名称常量和DB的列名。可用前缀、后缀或正则表达式来筛选需要生成字段常量的message。
- [**protoc-gen-go-gorm**](https://github.com/ml444/gkit/tree/master/cmd/protoc-gen-go-gorm): 生成gorm的model，通过proto扩展的方式定义了字段的gorm tag，以及json和bytes类型字段的序列化和反序列化方法。
- [**protoc-gen-go-http**](https://github.com/ml444/gkit/tree/master/cmd/protoc-gen-go-http): 生成http的路由和handler，并且可用pluck扩展模块实现http的请求头设置，这在上传和下载文件中比较有用。
- [**protoc-gen-go-validate**](https://github.com/ml444/gkit/tree/master/cmd/protoc-gen-go-validate): 生成参数校验的方法，可以根据`v.proto`定义的规则来校验参数。调用中间件`validation.Validator`
  来启动检验。

**安装protoc插件**

```shell
$ go get -u github.com/ml444/gkit/cmd/protoc-gen-go-errcode \
    github.com/ml444/gkit/cmd/protoc-gen-go-field \
    github.com/ml444/gkit/cmd/protoc-gen-go-gorm \
    github.com/ml444/gkit/cmd/protoc-gen-go-http \
    github.com/ml444/gkit/cmd/protoc-gen-go-validate
```

## <a name="complete_example"></a>完整proto示例

**protoc插件的使用示例**：

```protobuf
syntax = "proto3";

package user;

option go_package = "pkg/user";

import "v/v.proto";
import "err/err.proto";
import "orm/orm.proto";
import "pluck/pluck.proto";
import "dbx/pagination/pagination.proto";
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
            default_headers: {          // 设置默认的请求头
                content_type: "application/octet-stream"   
            }
            headers_to: "file_info"     // 将请求头提取到file_info结构体中
        };
    };

    rpc Download (DownloadReq) returns (DownloadRsp){
        option (google.api.http) = {
            post: "/v1/storage/download"
            body: "*"
        };
        option (pluck.response) = {
            default_headers: {          // 设置默认的响应头
                content_type: "application/vnd.openxmlformats"
                access_control_expose_headers: "Content-Disposition"
            }
            headers_from: "headers"     // 将headers结构体中的字段设置到http响应头中
            body_from: "data"           // 将data字段设置到http响应体中
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
    ModelUser data = 1 [(v.rules).message.required = true];     // 验证该结构体必须传值
}
message CreateUserRsp {
    ModelUser data = 1;
}

message UpdateUserReq {
    uint64 id = 1       [(v.rules).uint64.gt = 0];              // 验证ID值必须大于0
    ModelUser data = 2  [(v.rules).message.required = true];    // 验证ModelUser必须传值
}
message UpdateUserRsp {
    ModelUser data = 1;
}

message DeleteUserReq {
    uint64 id = 1                   [(v.rules).uint64.gte = 1];             // 验证id大于0
    // repeated uint64 id_list = 2  [(v.rules).repeated.min_items = 1];     // 验证最少传入一个ID
}
message DeleteUserRsp {}

message GetUserReq {
    uint64 id = 1       [(v.rules).uint64.gte = 1];          // 验证ID值必须大于等于1
}
message GetUserRsp {
    ModelUser data = 1;
}

message ListUserReq {
    repeated uint64 id_list = 1 [(v.rules).repeated.unique = true];                 // 校验数组内部的元素不能重复
    optional string name = 2    [(v.rules).string = {min_len: 1, max_len: 50}];     // 校验字符串长度大于等于1，小于等于50
    optional string phone = 3   [(v.rules).string = {pattern: "\\d+", min_len:6, max_len: 25}];     // 校验字符串长度大于等于6，小于等于25，并且符合正则表达式
    optional string email = 4   [(v.rules).string.email = true];        // 校验是否是邮箱的格式
    pagination.Pagination pagination = 5;
}
message ListUserRsp {
    pagination.Pagination pagination = 1;
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
import "v/v.proto";
import "err/err.proto";
import "orm/orm.proto";
import "pluck/pluck.proto";
import "dbx/pagination/pagination.proto";
```

proto内部import是引用了`gctl-templates/protos/gkit`,
如果你把这些导入文件放在其他地方，你可以修改为`import "your/path/xxx.proto`

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
- **dbx**: 基于gorm进行二次封装的数据库模块，封装了查询的链式方法(Eq\In\Like...)，支持软删除和分页查询。
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
- 可以设置多语言错误信息，根据请求头的`Accept-Language`来返回对应的错误信息
- 根据自定义错误码判断是否为指定的错误码
- 根据GRPC的错误码转化为http状态码
- 根据http状态码转化为GRPC的错误码

可以通过在proto文件中定义错误码和错误信息，然后通过errorx.RegisterErrCode()注册错误码和错误信息，
  然后通过`errorx.New()`等方法实例化`Error`对象。在请求返回错误时会返回对应的http状态码和错误信息。

```protobuf
syntax = "proto3";

package user;

import "err/err.proto";     // 源文件: github.com/ml444/gkit/cmd/protoc-gen-go-errcode/err/err.proto

// range of error codes: [102000, 102999]
enum ErrCode {
    option (err.lower_bound) = 101000;
    option (err.upper_bound) = 101999;

    Success = 0;
    ErrIllegalParam = 102000 [(err.detail) = {status:400, message:"无效参数", polyglot: ["zh=无效参数", "en=Invalid parameters"]}];
    ErrParamRequired = 102001 [(err.detail) = {status:400, message:"缺失参数", polyglot: ["zh=缺失参数", "en=Missing parameters"]}];
    ErrNotFoundUser = 102002 [(err.detail) = {status:404, message:"未找到用户", polyglot: ["zh=未找到用户", "en=Record not found"]}];
    // or 不配置多语言信息，默认使用message的内容
    // ErrIllegalParam = 102000 [(err.detail) = {status:400, message:"非法参数"}];
    // ErrParamRequired = 102001 [(err.detail) = {status:400, message:"缺失参数"}];
    // ErrNotFoundUser = 102002 [(err.detail) = {status:404, message:"未找到用户"}];
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
		return nil, errorx.New(user.ErrParamRequired) 
        // 如果存在请求头: `Accept-Language: en` 
        // 返回英文的错误信息: {"status_code": 400, "error_code": 102001, "message": "Missing parameters"}
        // 如果存在请求头: `Accept-Language: zh` 
        // 返回中文的错误信息: {"status_code": 400, "error_code": 102001, "message": "缺失参数"}
	}
	// do something

	// if not found user
	return nil, errorx.New(user.ErrNotFoundUser) 
    // 如果存在请求头: `Accept-Language: en` 
    // 返回英文的错误信息: {"status_code": 404, "error_code": 102002, "message": "The user was not found"}
    // 如果存在请求头: `Accept-Language: zh` 
    // 返回中文的错误信息: {"status_code": 404, "error_code": 102002, "message": "未找到用户"}
}

func main() {
	errorx.RegisterError(user.ErrCodeMap)
	// pass
}
```

### dbx

基于gorm的数据库模块和Proto格式的model做了一些预处理的工作，主要包含以下几个功能：

- 封装gorm的增删改查，查询封装了链式方法(Eq\Gt\Lt\In\NotIn\Between...)，使其更易于使用，并支持软删除以及分页查询。
- 封装了复杂查询的参数结构`QueryOpts`, 在一些复杂的查询下，可以更方便地处理查询条件。
- 针对NotFoundRecord的错误处理，可以自定义错误码和错误信息。
- 封装列表的分页查询`dbx.pagination`，使分页查询更易于使用。

基础使用：

```go
package main

import (
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"github.com/ml444/gkit/dbx"
	"github.com/ml444/gkit/dbx/pagination"
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
	// Or use pagination query to get total count
	pag, err := scope.LikePrefix("name", "test").Lte("age", 25).PaginateQuery(&pagination.Pagination{Page: 1, Size: 10, SkipCount: false}, &users)
	// return pagination: Paginate{Total: 100, Page: 1, Size: 10} 

	// GroupBy and Having
	var userGroup []*GroupBy
	err = scope.Select("name", "count(*) AS total").Group("name").Having("age > 18").Find(&userGroup)

	// OrderBy
	err = scope.Order("age DESC").Find(&users)
}
```

#### 分页查询

分页查询的参数和结果返回，都是通过`pagination.Pagination`来定义的，可以根据实际情况来选择不同分页方式。
分页查询也有两种使用方式：

1. 通过指定页数和每页数量来查询。
2. 滚动翻页查询，这种方式适合于大数据量的查询。

_**注意**_：分页查询时可以在第二页之后调用`skip_count`参数来节约数据库的性能，当然这需要前端工程师缓存首次拿到的总数。

分页的数据库查询可以使用`Scope.PaginateQuery()`方法，内部调用了`Count()`和`Find()`。

**分页方式的proto定义**：

```protobuf
syntax = "proto3";
import "dbx/pagination/pagination.proto";
/*
message Paginate {
  uint32 page = 1;
  uint32 size = 2;
  int64 total = 5;
  bool skip_count = 6;
}
 */

message ListUserReq {
    pagination.Pagination pagination = 1;   // 指定页数和每页数量 
}

message ListUserRsp {
    pagination.Pagination pagination = 1;
}
```

**滚动方式的proto定义**：

```protobuf
syntax = "proto3";
import "dbx/pagination/pagination.proto";
message ListUserReq {
    pagination.Scroll scroll = 1;     // 滚动翻页查询
}
message ListUserRsp {
    repeated ModelUser list = 2;
}
```

### optx

过滤筛选查询模块，有两种方式（枚举和指针）来定义列表的过滤查询参数。而且都封装相应的处理器`Processor`和对应的处理参数的方法，规范化处理列表查询参数。

本模块封装了两种列表参数筛选查询方式，一种是通过`optx.Options`来定义查询参数，另一种是直接定义指针参数来查询。
具体使用哪种方式，可以根据实际情况来选择，如果是对外隐蔽查询参数的API，可以使用`optx.Options`来定义查询参数，
这样可以更好地控制查询参数，并对参数含义具备隐蔽性，因为代表其参数的是枚举值。
如果一般情况或者需要传递零值的场景，可以直接定义指针参数来查询，这种方式直接。

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
	err := optx.NewProcessor().
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
		Process(req.ListOption) // OR: ProcessOptions(req.ListOption)
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
	err := optx.NewProcessor().
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
		Process(req)  // OR: ProcessStruct(req)
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

[proto的ListUser示例](#complete_example)，包含了数据库查询，过滤筛选和分页功能：

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
	err := optx.NewProcessor().
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

日志模块，主要包含以下几个功能：

- 定义了日志接口，可以自定义日志实现， 默认使用`os.stdout`输出
- 自定义gorm的日志输出，与自定义的日志实现结合，一起输出指定位置。

直接使用：

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
可以搭配glog进行各种灵活的输出，如文件输出、流传输、Syslog、控制台标准输出等。
使用[glog](https://github.com/ml444/glog)的日志文件存储示例:

```go
package main

import (
	"github.com/ml444/gkit/log"
	glog "github.com/ml444/glog"
)

// InitLogger 简单配置
func InitLogger(debug bool) error {
    opts := []glog.OptionFunc{
        glog.SetLoggerName("serviceName"),
        glog.SetWorkerConfigs(glog.NewDefaultTextFileWorkerConfig("./logs")),
    }
    if debug {
        opts = append(opts, glog.SetLoggerLevel(glog.DebugLevel))
    }
	err := glog.InitLog(opts...)
	if err != nil {
		return err
	}
	log.SetLogger(glog.GetLogger())
	return nil
}

// InitLogger 详细配置：
func InitLogger() error {
  return glog.InitLog(
    glog.SetLoggerName("serviceName"),   // 可选
    glog.SetWorkerConfigs(
      glog.NewWorkerConfig(glog.InfoLevel, 1024).SetFileHandlerConfig(
        glog.NewDefaultFileHandlerConfig("logs").
          WithFileName("text_log").       // 另外指定文件名
          WithFileSize(1024*1024*1024).   // 1GB
          WithBackupCount(12).            // 保留的日志文件数量
          WithBulkSize(1024*1024).        // 批量写入硬盘的大小
          WithInterval(60*60).            // 日志按每小时滚动切割
          WithRotatorType(glog.FileRotatorTypeTimeAndSize),
      ).SetJSONFormatterConfig(
        glog.NewDefaultJSONFormatterConfig().WithBaseFormatterConfig(
          glog.NewDefaultBaseFormatterConfig().
            WithEnableHostname().       // 记录服务器的hostname
            WithEnableTimestamp().      // 记录时间戳
            WithEnablePid().            // 记录进程ID
            WithEnableIP(),             // 记录服务器IP
        ),
      ),
    ),
  )
}

func main() {
	// 初始化日志
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

中间件，目前开发了以下中间件：

- 通用中间件：用于处理响应中的空响应和统一错误输出的的通用中间件
- 日志中间件： 记录请求日志、响应日志、请求和响应日志
- 限流中间件： 用于限制请求的访问频率
- 恢复中间件： 用于恢复panic，并记录日志
- 跟踪中间件： 用于请求的链路追踪
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
	"github.com/ml444/gkit/transport/httpx"
)

func main() {
	// HTTP
	httpx.NewServer(
		httpx.Address(":5050"),
		httpx.Middleware(
			// 通用中间件: 处理空响应
			general.ReplaceEmptyResponse(struct {
				StatusCode int32
				ErrCode    int32
				Message    string
			}{200, 0, "success"}),
			// 通用中间件: 把错误统一成errorx.Error的结构
			general.WrapError(),
			// 日志中间件: 记录请求的入参和耗时
			logging.LogRequest(),
			// 限频中间件: 
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
			// 恢复中间件
			recovery.Recovery(),
			// 跟踪中间件
			trace.Server(),
			// 参数校验中间件
			validate.Validator(),

		),
	)
}
```

### transport

传输模块，主要包含以下几个功能：

- 把http请求核心逻辑转化为grpc的Service方法。
- 封装了http和grpc的中间件，统一中间件的接口。

### pkg

公共模块，包含了一些基础的工具或功能，主要包含以下几个模块：

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

