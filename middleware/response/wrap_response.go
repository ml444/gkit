package response

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"strings"

	"github.com/ml444/gkit/middleware"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/anypb"
)

func WrapResponse() middleware.Middleware {
	return func(handler middleware.ServiceHandler) middleware.ServiceHandler {
		return func(ctx context.Context, req any) (rsp any, err error) {
			rsp, err = handler(ctx, req)
			if err == nil {
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
	wrote  bool
}

func (w *bodyCaptureWriter) WriteHeader(statusCode int) {
	if !w.wrote {
		w.status = statusCode
		w.wrote = true
	}
}

func (w *bodyCaptureWriter) Write(b []byte) (int, error) {
	if !w.wrote {
		w.WriteHeader(http.StatusOK)
	}
	return w.body.Write(b)
}

func (w *bodyCaptureWriter) Flush() {
	if f, ok := w.ResponseWriter.(http.Flusher); ok {
		f.Flush()
	}
}

// acceptsJSON reports whether the client accepts JSON responses.
func acceptsJSON(h http.Header) bool {
	accept := h.Get("Accept")
	if accept == "" {
		return true
	}
	return strings.Contains(accept, "application/json")
}

// WrapHttpResponse wraps successful JSON responses into {code,message,data}.
func WrapHttpResponse() middleware.HttpMiddleware {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
			if !acceptsJSON(req.Header) {
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
				for k, vs := range cw.Header() {
					for _, v := range vs {
						w.Header().Add(k, v)
					}
				}
				if cw.wrote {
					w.WriteHeader(cw.status)
				}
				_, _ = w.Write(cw.body.Bytes())
				return
			}

			if cw.status < 200 || cw.status >= 300 {
				if cw.wrote {
					w.WriteHeader(cw.status)
				}
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
