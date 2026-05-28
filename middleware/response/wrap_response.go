package response

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"

	"github.com/ml444/gkit/middleware"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/anypb"
)

func WrapResponse() middleware.Middleware {
	return func(handler middleware.ServiceHandler) middleware.ServiceHandler {
		return func(ctx context.Context, req any) (rsp any, err error) {
			rsp, err = handler(ctx, req)
			if err == nil {
				// rsp = map[string]interface{}{
				// 	"code":    0,
				// 	"message": "success",
				// 	"data":    rsp,
				// }
				var value anypb.Any
				err = anypb.MarshalFrom(&value, rsp.(proto.Message), proto.MarshalOptions{})
				if err != nil {
					return nil, err
				}
				rsp = &ApiCommonResponse{
					Code:    0,
					Message: "success",
					Data:    &value,
				}
			}
			return rsp, err
		}
	}
}

type bodyCaptureWriter struct {
	http.ResponseWriter
	status int
	body   *bytes.Buffer
}

func (w *bodyCaptureWriter) WriteHeader(statusCode int) {
	w.status = statusCode
	w.ResponseWriter.WriteHeader(statusCode)
}

func (w *bodyCaptureWriter) Write(b []byte) (int, error) {
	return w.body.Write(b)
}

// WrapHttpResponse wraps successful JSON responses into {code,message,data}.
// For non-2xx responses, it leaves the response unchanged.
func WrapHttpResponse() middleware.HttpMiddleware {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
			if req.Header.Get("Accept") != "application/json" {
				next.ServeHTTP(w, req)
				return
			}

			cw := &bodyCaptureWriter{
				ResponseWriter: w,
				status:         http.StatusOK,
				body:           &bytes.Buffer{},
			}
			next.ServeHTTP(cw, req)

			if IsHttpRaw(cw.Header()) {
				_, _ = w.Write(cw.body.Bytes())
				return
			}

			if cw.status < 200 || cw.status >= 300 {
				_, _ = w.Write(cw.body.Bytes())
				return
			}

			original := cw.body.Bytes()
			var data any
			if len(original) == 0 {
				data = nil
			} else if err := json.Unmarshal(original, &data); err != nil {
				data = string(original)
			}

			wrapped := map[string]any{
				"code":    0,
				"message": "success",
				"data":    data,
			}

			w.Header().Set("Content-Type", "application/json; charset=utf-8")
			_ = json.NewEncoder(w).Encode(wrapped)
		})
	}
}
