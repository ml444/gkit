# gkit
gkit是一个集成`HTTP`和`GRPC`通信协议的微服务框架，主旨在于易使用，通过封装日常WEB开发的组件，
配合`gctl`工具，可以快速的生成项目的基础代码，从而开发人员可以更专注于业务逻辑开发。

主要包含以下几个模块：

## 项目结构介绍

```
.
├── dbx
├── errorx
├── listoption     
├── log
├── metrics
├── middleware
├── pkg
│   ├── auth
│   ├── env
│   ├── header
│   └── routine
├── transport
├── go.mod
└── go.sum
```
- **errorx**: 错误处理模块
- **dbx**: 基于gorm的数据库模块
- **listoption**: 分页查询模块，定义了列表数据的条件筛选方式，以及分页查询的参数和结果返回。
- **log**: 日志模块
- **metrics**: 指标模块
- **middleware**: 中间件模块
- **transport**: 传输模块
- **pkg**: 公共模块，包含了一些基础的工具类
- **internal**: 内部模块

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
enum ErrCode {
    Success = 0;
    // @status_code: 400
    ErrInvalidParam = 102001;
    // @status_code: 400
    ErrParamRequired = 102002; // 缺失参数
    // @status_code: 404
    ErrNotFoundUser = 102003; // 未找到用户
    // @status_code: 500
    ErrCreateUserFailed = 102004; // 创建用户失败
    // @status_code: 403
    ErrUserExisted = 102005; // 用户已存在
}
```

```go
package main

func (s *UserService) GetUser(ctx context.Context, req *pb.GetUserReq) (*pb.GetUserRsp, error) {
	return nil, errorx.New(ErrNotFoundUser)
	// return: {"status_code": 404, "error_code": 102003, "message": "未找到用户"}
}
```

### listoption

列表查询模块，主要包含以下几个功能：

- 封装列表查询参数和规范化处理列表查询参数
- 封装列表的分页查询

本模块封装了两种列表参数筛选查询方式，一种是通过`listoption.Options`来定义查询参数，另一种是通过直接定义指针参数来查询。
具体使用哪种方式，可以根据实际情况来选择，如果是对外提供的接口，建议使用`listoption.Options`来定义查询参数，这样可以更好的控制查询参数，并具备隐蔽性。
如果是内部接口，可以直接定义指针参数来查询，这样可以更方便的使用。
分页查询的参数和结果返回，都是通过`listoption.Paginate`来定义的，可以根据实际情况来选择是否使用分页查询。
分页查询也有三种使用方式：

1. 通过指定页数和每页数量来查询，这种方式适合于前端分页查询。
2. 通过指定偏移量和每页数量来查询，这种方式适合于后端分页查询。
3. 滚动翻页查询，这种方式适合于大数据量的查询。
   分页查询也可以通过`dbx.PaginateQuery()`来实现，这样可以更方便的使用。

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
    // @ref_to: ListUserSysReq.ListOpt
    listoption.Options list_option = 1;
    listoption.Paginate paginate = 2;
}
message ListUserRsp {
    listoption.Paginate paginate = 1;
    repeated ModelUser list = 2;
}
```

```protobuf
syntax = "proto3";
import "optx/optx.proto";
message ListUserReq {
    repeated uint64 id_list = 1;
    optional string like_name = 2;
    optional string phone = 3;
    listoption.Paginate paginate = 4;
}
message ListUserRsp {
    listoption.Paginate paginate = 1;
    repeated ModelUser list = 2;
}
```

```protobuf
syntax = "proto3";
import "optx/optx.proto";
message ListUserReq1 {
    listoption.Paginate paginate = 1;   // 指定页数和每页数量 或 指定偏移量和每页数量
}
message ListUserReq2 {
    listoption.Scroll scroll = 1;   // 滚动翻页查询
}
```

### dbx

基于gorm的数据库模块和Proto格式的model做了一些预处理的工作，主要包含以下几个功能：

- 初始化数据库并设置日志和最大连接数和最大空闲连接数和最大生存时间和最大空闲时间
- 封装gorm的增删改查，使其更易于使用，并支持软删除以及分页查询。
- 针对NotFoundRecord的错误处理，可以自定义错误码和错误信息。

```go
package main

import (
	"github.com/ml444/gkit/dbx"
	"github.com/ml444/gkit/listoption"
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

func main() {
	var err error
	var scope = dbx.NewScope(dbx.Db(), &ModelUser{})
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
	// Or
	paginate, err := scope.LikePrefix("name", "test").Lte("age", 25).PaginateQuery(&listoption.Paginate{Page: 1, Size: 10}, &users)
	// return paginate: Paginate{Total: 100, Page: 1, Size: 10} 

	// GroupBy and Having
	var userGroup []*GroupBy
	err = scope.Select("name", "count(*) AS total").Group("name").Having("age > 18").Find(&userGroup)

	// OrderBy
	err = scope.Order("age DESC").Find(&users)
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

### metrics

指标模块，主要包含以下几个功能：

- 初始化指标

### middleware

中间件模块，主要包含以下几个功能：

- 记录请求日志
- 记录响应日志
- 记录请求和响应日志

### transport

传输模块，主要包含以下几个功能：

- http传输
    - 把http请求转化为grpc请求
    - 根据传入的grpc Service生成http路由
    - http请求头的转换为context内容
    - 前置中间件
- grpc传输

### validator

参数校验模块，主要包含以下几个功能：

### pkg

公共模块，包含了一些基础的工具类，主要包含以下几个模块：

- **auth**: 基于jwt的认证模块，主要包含以下几个功能：
    - 生成/解析/验证`access token`和`refresh token`
    - 设置过期时间
    - 自定义claim
- **env**: 用于工作环境的判断，如：是否为开发环境、测试环境、生产环境等
- **header**: 请求头相关的处理及httpHeader与Context的转换
- **routine**: 用于协程的安全处理，在协程中使用`routine.Go()`代替`go`关键字，可以捕获协程中的panic，并记录日志。

## proto 使用示例

```protobuf
syntax = "proto3";

package user;

option go_package = "gitlab.xxx.com/group1/project1/pkg/user";

import "validate/validate.proto";
import "optx/optx.proto";
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
            body: "*"
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
}

enum ErrCode {
    Success = 0;
    // @status_code: 400
    ErrInvalidParam = 102001;
    // @status_code: 400
    ErrParamRequired = 102002; // 缺失参数
    // @status_code: 404
    ErrNotFoundUser = 102003; // 未找到用户
    // @status_code: 500
    ErrCreateUserFailed = 102004; // 创建用户失败
    // @status_code: 403
    ErrUserExisted = 102005; // 用户已存在
}

message ModelUser {
    // @gorm: primarykey
    uint64 id = 1;
    uint32 created_at = 2;
    uint32 updated_at = 3;
    uint32 deleted_at = 4;
    uint32 sex = 5;
    // @gorm: type:varchar(40)
    string name = 6;
    // @gorm: type:varchar(40);uniqueIndex:idx_phone
    string phone = 7;
    // @gorm: type:varchar(255)
    string email = 8;
    // @gorm: type:varchar(255)
    string avatar = 9;
}

message CreateUserReq {
    ModelUser user = 1 [(validate.rules).message.required = true];
}
message CreateUserRsp {
    ModelUser user = 1;
}

message CreateUserSysReq {
    ModelUser user = 1 [(validate.rules).message.required = true];
}
message CreateUserSysRsp {
    ModelUser user = 1;
}

message UpdateUserReq {
    ModelUser user = 1 [(validate.rules).message.required = true];
}
message UpdateUserRsp {
    ModelUser user = 1;
}

message DeleteUserReq {
    uint64 id = 1 [(validate.rules).uint64.gt = 0];
}
message DeleteUserRsp {}

message GetUserReq {
    uint64 id = 1 [(validate.rules).uint64.gt = 0];
}
message GetUserRsp {
    ModelUser user = 1;
}

message GetUserByPhoneSysReq {
    string phone = 1[(validate.rules).string = {len:11}];
}
message  GetUserByPhoneSysRsp {
    ModelUser user = 1;
}

message ListUserReq {
    enum ListOpt {
        ListOptNil = 0;
        // @valueType: uint64List
        ListOptIdList = 1;
        // @valueType: string
        ListOptLikeName = 2;
        // @valueType: string
        ListOptPhone = 3;
    }
    // @ref_to: ListUserSysReq.ListOpt
    listoption.Options list_option = 1;
    listoption.Paginate paginate = 2;
}
message ListUserRsp {
    listoption.Paginate paginate = 1;
    repeated ModelUser list = 2;
}


message ListUserSysReq {
    enum ListOpt {
        ListOptNil = 0;
        // @valueType: uint64List
        ListOptIdList = 1;
        // @valueType: string
        ListOptLikeNickName = 2;
        // @valueType: string
        ListOptLikeRealName = 3;
        // @valueType: string
        ListOptPhone = 4;
        // @valueType: string
        ListOptEmail = 5;
    }
    listoption.ListOption list_option = 2;
}
message ListUserSysRsp {
    listoption.Paginate paginate = 1;
    repeated ModelUser list = 2;
}

message GetBaseUserInfoMapSysReq {
    uint64 corp_id = 1 [(validate.rules).uint64.gt = 0];
    repeated uint64 corp_user_id_list = 2 [(validate.rules).repeated.min_items = 1];
}
message GetBaseUserInfoMapSysRsp {
    // @desc: key: user_id
    map<uint64, BaseUserInfo> base_user_info_map = 1;
}

```

