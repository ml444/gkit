package transport

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"reflect"
	"strings"
	"time"

	"github.com/gorilla/mux"
	"github.com/ml444/gkit/auth"
	"github.com/ml444/gkit/errorx"
	"github.com/ml444/gkit/log"
	"github.com/ml444/gutil/str"
)

const (
	GET    = "GET"
	POST   = "POST"
	PUT    = "PUT"
	DELETE = "DELETE"
)

const contentType = "application/json; charset=utf-8"
const globalTimeout = 5 * time.Second

type HttpHandleFunc func(writer http.ResponseWriter, request *http.Request)
type callWithReflect func(in []reflect.Value) []reflect.Value
type validator interface {
	Validate() error
}

type EndpointParser struct {
	svc            interface{}
	router         *mux.Router
	timeoutMap     map[string]time.Duration
	jwtSecret      []byte
	jwtHook        auth.HookFunc
	enableCheckJWT bool
}

func (p *EndpointParser) Parse() error {
	var err error
	svcT := reflect.TypeOf(p.svc)
	svcV := reflect.ValueOf(p.svc)
	if !strings.HasSuffix(svcT.Name(), "Service") {
		err = fmt.Errorf("not found the suffix of 'Service' by %s", svcT.Name())
		log.Error(err.Error())
		return err
	}
	svcNamePrefix := str.ToLowerFirst(strings.TrimSuffix(svcT.Name(), "Service"))
	n := svcT.NumMethod()
	for i := 0; i < n; i++ {
		var httpMethod string
		mn := svcT.Method(i)
		funcName := mn.Name
		if d := funcName[0]; d <= 'A' || d >= 'Z' {
			continue
		}
		if strings.HasPrefix(funcName, "Get") || strings.HasPrefix(funcName, "List") {
			httpMethod = GET
		} else if strings.HasPrefix(funcName, "Create") {
			httpMethod = POST
		} else if strings.HasPrefix(funcName, "Update") {
			httpMethod = PUT
		} else if strings.HasPrefix(funcName, "Delete") {
			httpMethod = DELETE
		} else {
			httpMethod = POST
		}

		var timeout = globalTimeout
		if v, ok := p.timeoutMap[funcName]; ok {
			timeout = v
		}
		var req = reflect.New(mn.Type.In(2).Elem())

		p.router.Methods(httpMethod).PathPrefix("/" + svcNamePrefix).Path("/" + funcName).HandlerFunc(
			p.handleWithReflect(svcV, req, mn.Func.Call, timeout),
		)
	}
	return nil
}

func (p *EndpointParser) handleWithReflect(svcV, req reflect.Value, callFunc callWithReflect, timeout time.Duration) HttpHandleFunc {
	return func(writer http.ResponseWriter, request *http.Request) {
		defer func() {
			if Err := recover(); Err != nil {
				//var brokenPipe bool
				//if ne, ok := err.(*net.OpError); ok {
				//	var se *os.SyscallError
				//	if errorx.As(ne, &se) {
				//		if strings.Contains(strings.ToLower(se.Error()), "broken pipe") || strings.Contains(strings.ToLower(se.Error()), "connection reset by peer") {
				//			brokenPipe = true
				//		}
				//	}
				//}
				log.Errorf("%v", Err)
				writer.WriteHeader(http.StatusInternalServerError)
			}
		}()
		var (
			ctx    context.Context
			cancel context.CancelFunc
		)
		if timeout > 0 {
			ctx, cancel = context.WithTimeout(request.Context(), timeout)
		} else {
			ctx, cancel = context.WithCancel(request.Context())
		}
		defer cancel()

		writer.Header().Set("Content-Type", contentType)

		var err error
		var result interface{}
		if p.enableCheckJWT {
			err = auth.ParseJWT2ContextByHTTP(ctx, request, p.jwtSecret, p.jwtHook)
			if err != nil {
				log.Error(err)
				if Err, ok := err.(*errorx.Error); ok {
					writer.WriteHeader(int(Err.StatusCode))
					result = err
				} else {
					writer.WriteHeader(http.StatusInternalServerError)
					result = errorx.CreateError(errorx.UnknownStatusCode, errorx.ErrCodeInvalidHeaderSys, err.Error())
				}
				goto RETURN
			}
		}

		{
			var r = req.Interface()
			err = json.NewDecoder(request.Body).Decode(r)
			if err != nil && err != io.EOF {
				log.Errorf("err: %v", err)
				result = errorx.CreateError(errorx.UnknownStatusCode, errorx.ErrCodeInvalidReqSys, err.Error())
				goto RETURN
			}
			if v, ok := r.(validator); ok {
				if err = v.Validate(); err != nil {
					writer.WriteHeader(errorx.DefaultStatusCode)
					result = errorx.CreateError(errorx.DefaultStatusCode, errorx.ErrCodeInvalidParamSys, err.Error())
					goto RETURN
				}
			}

			values := callFunc([]reflect.Value{svcV, reflect.ValueOf(ctx), req})
			rspV := values[0]
			rspErr := values[1]
			if IErr := rspErr.Interface(); IErr != nil {
				if Err, ok := IErr.(*errorx.Error); ok {
					writer.WriteHeader(int(Err.StatusCode))
					result = IErr
				} else {
					writer.WriteHeader(http.StatusInternalServerError)
					result = errorx.CreateError(errorx.UnknownStatusCode, errorx.ErrCodeInvalidReqSys, err.Error())
				}
			} else {
				result = rspV.Interface()
			}
		}

	RETURN:
		var bodyBuf []byte
		bodyBuf, err = json.Marshal(result)
		if err != nil {
			log.Errorf("err: %v", err)
			return
		}
		_, err = writer.Write(bodyBuf)
		if err != nil {
			log.Errorf("err: %v", err)
			return
		}
	}
}

func NewEndpointParser(svc interface{}, router *mux.Router, opts ...OptionFunc) *EndpointParser {
	parser := &EndpointParser{
		svc:    svc,
		router: router,
	}
	for _, optFunc := range opts {
		optFunc(parser)
	}
	return parser
}

func ParseService2HTTP(svc interface{}, router *mux.Router, opts ...OptionFunc) error {
	parser := NewEndpointParser(svc, router, opts...)
	return parser.Parse()
}

type OptionFunc func(parser *EndpointParser)

func SetTimeoutMap(timeoutMap map[string]time.Duration) OptionFunc {
	return func(parser *EndpointParser) {
		parser.timeoutMap = timeoutMap
	}
}

func SetJwtSecret(secret []byte) OptionFunc {
	return func(parser *EndpointParser) {
		parser.jwtSecret = secret
		parser.enableCheckJWT = true
	}
}

func SetJwtHook(hook auth.HookFunc) OptionFunc {
	return func(parser *EndpointParser) {
		parser.jwtHook = hook
	}
}
