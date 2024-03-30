package account

const (
	Success                    int32 = 0
	ErrInvalidParam            int32 = 101001
	ErrEmptyParams             int32 = 101002
	ErrAccountLength           int32 = 101003
	ErrNotFoundAccount         int32 = 101004
	ErrPasswordLength          int32 = 101005
	ErrPasswordConfirm         int32 = 101006
	ErrPasswordFailed          int32 = 101007
	ErrNotFoundToken           int32 = 101008
	ErrTokenConfirm            int32 = 101009
	ErrLoginToken              int32 = 101010
	ErrAccountExisted          int32 = 101011
	ErrAccountHasNotJoinedCorp int32 = 101012
)

var ErrCodeMap = map[int32]string{
	Success:                    "Success",
	ErrInvalidParam:            "ErrInvalidParam",
	ErrEmptyParams:             "ErrEmptyParams",
	ErrAccountLength:           "账户长度必须在6到20个字符之间",
	ErrNotFoundAccount:         "未找到账号",
	ErrPasswordLength:          "密码长度必须在6到20个字符之间",
	ErrPasswordConfirm:         "两次密码不一致",
	ErrPasswordFailed:          "密码错误",
	ErrNotFoundToken:           "未找到Token",
	ErrTokenConfirm:            "请求前后token不一致",
	ErrLoginToken:              "登录Token超时或错误",
	ErrAccountExisted:          "账号已存在",
	ErrAccountHasNotJoinedCorp: "账号没有加入该企业",
}

var ErrCode4StatusCodeMap = map[int32]int32{
	ErrInvalidParam:            400,
	ErrEmptyParams:             400,
	ErrAccountLength:           400,
	ErrNotFoundAccount:         404,
	ErrPasswordLength:          400,
	ErrPasswordConfirm:         400,
	ErrPasswordFailed:          401,
	ErrNotFoundToken:           403,
	ErrTokenConfirm:            403,
	ErrAccountExisted:          403,
	ErrAccountHasNotJoinedCorp: 403,
}
