package account

import (
	"context"
)

const ClientName = "account"
const ModuleCode uint32 = 0

func Register(ctx context.Context, req *RegisterReq) (*RegisterRsp, error) {
	if cliMgr.conn == nil {
		return nil, cliMgr.initErr
	}
	return cliMgr.cli.Register(ctx, req)
}

func Login(ctx context.Context, req *LoginReq) (*LoginRsp, error) {
	if cliMgr.conn == nil {
		return nil, cliMgr.initErr
	}
	return cliMgr.cli.Login(ctx, req)
}

func Logout(ctx context.Context, req *LogoutReq) (*LogoutRsp, error) {
	if cliMgr.conn == nil {
		return nil, cliMgr.initErr
	}
	return cliMgr.cli.Logout(ctx, req)
}

func Token(ctx context.Context, req *TokenReq) (*TokenRsp, error) {
	if cliMgr.conn == nil {
		return nil, cliMgr.initErr
	}
	return cliMgr.cli.Token(ctx, req)
}

func Refresh(ctx context.Context, req *RefreshReq) (*RefreshRsp, error) {
	if cliMgr.conn == nil {
		return nil, cliMgr.initErr
	}
	return cliMgr.cli.Refresh(ctx, req)
}

func VerifyCode(ctx context.Context, req *VerifyCodeReq) (*VerifyCodeRsp, error) {
	if cliMgr.conn == nil {
		return nil, cliMgr.initErr
	}
	return cliMgr.cli.VerifyCode(ctx, req)
}

func GetPublicKey(ctx context.Context, req *GetPublicKeyReq) (*GetPublicKeyRsp, error) {
	if cliMgr.conn == nil {
		return nil, cliMgr.initErr
	}
	return cliMgr.cli.GetPublicKey(ctx, req)
}

func ResetPassword(ctx context.Context, req *ResetPasswordReq) (*ResetPasswordRsp, error) {
	if cliMgr.conn == nil {
		return nil, cliMgr.initErr
	}
	return cliMgr.cli.ResetPassword(ctx, req)
}

func GetAccountInfo(ctx context.Context, req *GetAccountInfoReq) (*GetAccountInfoRsp, error) {
	if cliMgr.conn == nil {
		return nil, cliMgr.initErr
	}
	return cliMgr.cli.GetAccountInfo(ctx, req)
}

func ListAccount(ctx context.Context, req *ListAccountReq) (*ListAccountRsp, error) {
	if cliMgr.conn == nil {
		return nil, cliMgr.initErr
	}
	return cliMgr.cli.ListAccount(ctx, req)
}

func CreateAccountSys(ctx context.Context, req *CreateAccountSysReq) (*CreateAccountSysRsp, error) {
	if cliMgr.conn == nil {
		return nil, cliMgr.initErr
	}
	return cliMgr.cli.CreateAccountSys(ctx, req)
}

func DecryptPwdSys(ctx context.Context, req *DecryptPwdSysReq) (*DecryptPwdSysRsp, error) {
	if cliMgr.conn == nil {
		return nil, cliMgr.initErr
	}
	return cliMgr.cli.DecryptPwdSys(ctx, req)
}
