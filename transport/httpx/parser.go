package httpx

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/ml444/gkit/middleware"
	"github.com/ml444/gkit/middleware/httpmw"
	"io"
	"net/http"
	"reflect"
	"strings"
	"time"

	"github.com/gorilla/mux"
	"github.com/ml444/gkit/auth/jwt"
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
type ReflectCallFunc func(in []reflect.Value) []reflect.Value
type validator interface {
	Validate() error
}

type EndpointParser struct {
	svc            interface{}
	router         *mux.Router
	timeoutMap     map[string]time.Duration
	jwtHook        jwt.HookFunc
	enableCheckJWT bool

	beforeHandlerList []middleware.BeforeHandler
	afterHandlerList  []middleware.AfterHandler
}

func NewEndpointParser(svc interface{}, router *mux.Router, opts ...OptionFunc) *EndpointParser {
	parser := &EndpointParser{
		svc:               svc,
		router:            router,
		beforeHandlerList: []middleware.BeforeHandler{httpmw.Validator()},
		//afterHandlerList:  []middleware.AfterHandler{httpmw.CheckResponseError()},
	}
	for _, optFunc := range opts {
		optFunc(parser)
	}
	return parser
}

func (p *EndpointParser) Parse() error {
	var err error
	svcT := reflect.TypeOf(p.svc)
	if !strings.HasSuffix(svcT.Name(), "Service") {
		err = fmt.Errorf("not found the suffix of 'Service' by %s", svcT.Name())
		log.Error(err.Error())
		return err
	}
	svcNamePrefix := str.ToLowerFirst(strings.TrimSuffix(svcT.Name(), "Service"))
	n := svcT.NumMethod()
	for i := 0; i < n; i++ {
		var httpMethod = POST
		mn := svcT.Method(i)
		funcName := mn.Name
		if d := funcName[0]; d <= 'A' || d >= 'Z' {
			continue
		}
		if strings.HasSuffix(funcName, "Sys") {
			continue
		}
		//if strings.HasPrefix(funcName, "Get") || strings.HasPrefix(funcName, "List") {
		//	httpMethod = GET
		//} else if strings.HasPrefix(funcName, "Create") {
		//	httpMethod = POST
		//} else if strings.HasPrefix(funcName, "Update") {
		//	httpMethod = PUT
		//} else if strings.HasPrefix(funcName, "Delete") {
		//	httpMethod = DELETE
		//} else {
		//	httpMethod = POST
		//}

		var timeout = globalTimeout
		if v, ok := p.timeoutMap[funcName]; ok {
			timeout = v
		}
		var req = reflect.New(mn.Type.In(2).Elem())

		p.router.Methods(httpMethod).PathPrefix("/" + svcNamePrefix).Path("/" + funcName).HandlerFunc(
			p.handleWithReflect(req, mn.Func.Call, timeout),
		)
	}
	return nil
}

func (p *EndpointParser) handleWithReflect(req reflect.Value, callFunc ReflectCallFunc, timeout time.Duration) HttpHandleFunc {
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
				log.Fatalf("%v", Err)
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
		//ctx = context.WithValue(ctx, core.KeyHeaders, request.Header)
		var err error
		var rspResult interface{}
		//if p.enableCheckJWT
		{
			ctx, err = HandleContextByHTTP(ctx, request, p.jwtHook)
			if err != nil {
				log.Error(err)
				if Err, ok := err.(*errorx.Error); ok {
					writer.WriteHeader(int(Err.StatusCode))
					rspResult = err
				} else {
					writer.WriteHeader(http.StatusInternalServerError)
					rspResult = errorx.CreateError(errorx.UnknownStatusCode, errorx.ErrCodeInvalidHeaderSys, err.Error())
				}
				goto RETURN
			}
		}

		{
			var r = req.Interface()
			err = json.NewDecoder(request.Body).Decode(r)
			if err != nil && err != io.EOF {
				log.Errorf("err: %v", err)
				rspResult = errorx.CreateError(errorx.UnknownStatusCode, errorx.ErrCodeInvalidReqSys, err.Error())
				goto RETURN
			}
			log.Debugf("req[%s]: %v", req.Type().Elem().Name(), r)
			// processing before handler
			for _, h := range p.beforeHandlerList {
				ctx, r, err = h(ctx, r)
				if err != nil {
					log.Error(err)
					if Err, ok := err.(*errorx.Error); ok {
						writer.WriteHeader(int(Err.StatusCode))
						rspResult = err
					} else {
						writer.WriteHeader(http.StatusInternalServerError)
						rspResult = errorx.CreateError(errorx.UnknownStatusCode, errorx.ErrCodeInvalidBodySys, err.Error())
					}
					goto RETURN
				}

			}

			svcV := reflect.ValueOf(p.svc)
			values := callFunc([]reflect.Value{svcV, reflect.ValueOf(ctx), req})
			rspResult = values[0].Interface()
			rspErr := values[1].Interface()
			for _, h := range p.afterHandlerList {
				rspResult, err = h(rspResult, rspErr.(error))
				if err != nil {
					log.Errorf("rsp err: %v", err)
					if Err, ok := err.(*errorx.Error); ok {
						writer.WriteHeader(int(Err.StatusCode))
						rspResult = err
					} else {
						writer.WriteHeader(http.StatusInternalServerError)
						rspResult = errorx.CreateError(errorx.UnknownStatusCode, errorx.ErrCodeInvalidBodySys, err.Error())
					}
					goto RETURN
				}
			}
		}

	RETURN:
		var bodyBuf []byte
		bodyBuf, err = json.Marshal(rspResult)
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

func (p *EndpointParser) WithOptions(opts ...OptionFunc) {
	for _, optFunc := range opts {
		optFunc(p)
	}
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

func SetJwtHook(hook jwt.HookFunc) OptionFunc {
	return func(parser *EndpointParser) {
		parser.jwtHook = hook
	}
}

func AddBeforeHandler(handler middleware.BeforeHandler) OptionFunc {
	return func(parser *EndpointParser) {
		parser.beforeHandlerList = append(parser.beforeHandlerList, handler)
	}
}

func AddAfterHandler(handler middleware.AfterHandler) OptionFunc {
	return func(parser *EndpointParser) {
		parser.afterHandlerList = append(parser.afterHandlerList, handler)
	}
}
