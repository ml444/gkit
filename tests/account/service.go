package account

import (
	"context"
	"fmt"

	"github.com/ml444/gkit/errorx"
)

type AccountService struct {
	UnsafeAccountServer
}

func NewAccountService() AccountService {
	return AccountService{}
}

func (s AccountService) Register(ctx context.Context, req *RegisterReq) (*RegisterRsp, error) {
	var rsp RegisterRsp
	if req.Account == "" || req.Password == "" {
		return nil, errorx.New(ErrEmptyParams)
	}

	fmt.Printf("===> %v \n", req)
	rsp.Success = true
	rsp.Msg = "register successfully"
	return &rsp, nil
}

func (s AccountService) Login(ctx context.Context, req *LoginReq) (*LoginRsp, error) {
	var rsp LoginRsp
	if req.Account == "" || req.Password == "" {
		return nil, errorx.CreateError(400, ErrEmptyParams, "account or password is empty")
	}

	return &rsp, nil
}

func (s AccountService) Logout(ctx context.Context, req *LogoutReq) (*LogoutRsp, error) {
	return &LogoutRsp{}, nil
}

func (s AccountService) Token(ctx context.Context, req *TokenReq) (*TokenRsp, error) {
	var rsp TokenRsp
	if req.CorpId == 0 || req.LoginToken == "" {
		return nil, errorx.CreateError(400, ErrEmptyParams, "corp or token is empty")
	}

	return &rsp, nil
}

func (s AccountService) Refresh(ctx context.Context, req *RefreshReq) (*RefreshRsp, error) {
	return &RefreshRsp{}, nil
}

func (s AccountService) VerifyCode(ctx context.Context, req *VerifyCodeReq) (*VerifyCodeRsp, error) {
	var rsp VerifyCodeRsp

	return &rsp, nil
}

func (s AccountService) GetPublicKey(ctx context.Context, req *GetPublicKeyReq) (*GetPublicKeyRsp, error) {
	var rsp GetPublicKeyRsp
	return &rsp, nil
}

func (s AccountService) ResetPassword(ctx context.Context, req *ResetPasswordReq) (*ResetPasswordRsp, error) {
	var rsp ResetPasswordRsp

	rsp.Success = true
	rsp.Msg = "reset successfully"
	return &rsp, nil
}
func (s AccountService) GetAccountInfo(ctx context.Context, req *GetAccountInfoReq) (*GetAccountInfoRsp, error) {
	var rsp GetAccountInfoRsp
	fmt.Println("===> hello: ", req.Account)
	return &rsp, nil
}

func (s AccountService) ListAccount(ctx context.Context, req *ListAccountReq) (*ListAccountRsp, error) {
	var rsp ListAccountRsp
	fmt.Printf("===> %v \n", req)
	return &rsp, nil
}

func (s AccountService) CreateAccountSys(ctx context.Context, req *CreateAccountSysReq) (*CreateAccountSysRsp, error) {
	var rsp CreateAccountSysRsp
	fmt.Printf("===> %v \n", req)
	return &rsp, nil
}

func (s AccountService) DecryptPwdSys(ctx context.Context, req *DecryptPwdSysReq) (*DecryptPwdSysRsp, error) {
	var rsp DecryptPwdSysRsp

	return &rsp, nil
}
