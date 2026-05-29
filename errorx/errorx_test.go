package errorx

import (
	"errors"
	"net/http"
	"sync"
	"testing"

)

func TestGeneralHTTPStatus(t *testing.T) {
	t.Parallel()
	cases := []struct {
		name   string
		err    *Error
		status int
		is     func(error) bool
	}{
		{"ServiceUnavailable", ServiceUnavailable("unavailable"), http.StatusServiceUnavailable, IsServiceUnavailable},
		{"GatewayTimeout", GatewayTimeout("timeout"), http.StatusGatewayTimeout, IsGatewayTimeout},
		{"ClientClosed", ClientClosed("closed"), StatusClientClosed, IsClientClosed},
	}
	for _, tc := range cases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			if int(tc.err.Status) != tc.status {
				t.Fatalf("status = %d, want %d", tc.err.Status, tc.status)
			}
			if !tc.is(tc.err) {
				t.Fatalf("is helper returned false for %s", tc.name)
			}
		})
	}
}

func TestGRPCRoundTrip(t *testing.T) {
	t.Parallel()
	orig := &Error{
		ErrorInfo: ErrorInfo{
			Status:  http.StatusBadRequest,
			Code:    102001,
			Message: "missing parameter",
			Metadata: map[string]string{
				"field": "id",
			},
		},
	}
	got := FromError(orig.GRPCStatus().Err())
	if got.Code != orig.Code {
		t.Fatalf("code = %d, want %d", got.Code, orig.Code)
	}
	if got.Status != orig.Status {
		t.Fatalf("status = %d, want %d", got.Status, orig.Status)
	}
	if got.Message != orig.Message {
		t.Fatalf("message = %q, want %q", got.Message, orig.Message)
	}
	if got.Metadata["field"] != "id" {
		t.Fatalf("metadata = %v", got.Metadata)
	}
}

func TestFromErrorPreservesCause(t *testing.T) {
	t.Parallel()
	root := errors.New("root cause")
	got := FromError(root)
	if got == nil {
		t.Fatal("expected non-nil error")
	}
	if !errors.Is(got, root) {
		t.Fatalf("unwrap chain broken: %v", got.Unwrap())
	}
}

func TestRegisterErrorDoesNotMutateCallerMap(t *testing.T) {
	detail := &ErrCodeDetail{
		Status:  0,
		Message: "msg",
		Code:    99,
	}
	m := map[int32]*ErrCodeDetail{99: detail}
	RegisterError(m)
	if detail.Status != 0 {
		t.Fatalf("caller detail mutated: status=%d", detail.Status)
	}
	lock.RLock()
	registered := errCodeMap[99]
	lock.RUnlock()
	if registered.Status != DefaultStatusCode {
		t.Fatalf("registered status = %d, want %d", registered.Status, DefaultStatusCode)
	}
}

func TestConvertMsgByLang(t *testing.T) {
	const code int32 = 88001
	RegisterError(map[int32]*ErrCodeDetail{
		code: {
			Status:  400,
			Code:    code,
			Message: "默认",
			Polyglot: map[string]string{
				"en": "default en",
				"zh": "默认中文",
			},
		},
	})
	SetLang("zh")
	e := New(code)
	e.ConvertMsgByLang("en")
	if e.Message != "default en" {
		t.Fatalf("message = %q, want %q", e.Message, "default en")
	}
	e2 := New(code)
	before := e2.Message
	e2.ConvertMsgByLang("zh")
	if e2.Message != before {
		t.Fatalf("message = %q, want unchanged when Accept-Language matches SetLang (%q)", e2.Message, before)
	}
}

func TestErrorIsByCode(t *testing.T) {
	t.Parallel()
	e := CreateError(400, 42, "x")
	if !ErrorIs(e, 42) {
		t.Fatal("ErrorIs should match code")
	}
	if ErrorIs(e, 43) {
		t.Fatal("ErrorIs should not match different code")
	}
}

func TestErrorsIsStatusAndCode(t *testing.T) {
	t.Parallel()
	a := CreateError(400, 1, "a")
	b := CreateError(400, 1, "b")
	if !errors.Is(a, b) {
		t.Fatal("same status+code should match via Error.Is")
	}
}

func TestConcurrentRegisterAndNew(t *testing.T) {
	const code int32 = 77001
	var wg sync.WaitGroup
	for i := 0; i < 8; i++ {
		wg.Add(1)
		go func(n int) {
			defer wg.Done()
			RegisterError(map[int32]*ErrCodeDetail{
				code: {
					Status:  400,
					Code:    code,
					Message: "ok",
				},
			})
			_ = New(code)
		}(i)
	}
	wg.Wait()
}
