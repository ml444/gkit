package logging

import (
	"context"
	"encoding/json"
	"time"

	"github.com/ml444/gkit/log"
	"github.com/ml444/gkit/middleware"
	"github.com/ml444/gkit/pkg/header"
	"github.com/ml444/gkit/transport"
)

type Sanitizer interface {
	Sanitize() string
}

type (
	Took  struct{}
	Reply struct{}
)

// LogRequest logs request latency and payload after handler completes.
func LogRequest(fns ...middleware.LurkerFunc) middleware.Middleware {
	if len(fns) == 0 {
		fns = append(fns, defaultLogging())
	}
	return func(handler middleware.ServiceHandler) middleware.ServiceHandler {
		return func(ctx context.Context, req interface{}) (rsp interface{}, err error) {
			startTime := time.Now()
			rsp, err = handler(ctx, req)
			ctx = context.WithValue(ctx, Took{}, time.Since(startTime).Milliseconds())
			ctx = context.WithValue(ctx, Reply{}, rsp)
			if err != nil {
				log.Errorf("request failed: %v", err)
			}
			middleware.ForceLurkerChain(ctx, req, fns...)
			return
		}
	}
}

// 脱敏名单配置 (仅这些字段会打码，其他统统明文打印)
var sensitizedMap = map[string]bool{
	"pwd":      true,
	"password": true,
	"secret":   true,
}

func defaultLogging() middleware.LurkerFunc {
	return func(ctx context.Context, req any) error {
		took := ctx.Value(Took{})
		path := ""
		if tr, ok := transport.FromContext(ctx); ok {
			path = tr.Path()
		}
		ti := header.TraceInfoFromContext(ctx)
		trace := ti.TraceID
		if trace == "" {
			trace = header.CorrelationID(ctx)
		}
		// --- 对 req 进行脱敏处理 ---
		safeReq := sanitizeReq(req, sensitizedMap)
		log.Infof("trace=%s span=%s path=%s took=%vms req=%s", trace, ti.SpanID, path, took, safeReq)
		return nil
	}
}

// sanitizeReq 用于对请求数据进行脱敏处理
func sanitizeReq(req any, sensitizedMap map[string]bool) string {
	if req == nil {
		return ""
	}
	// 如果 Request 结构体实现一个自定义的 Sanitize() string 方法，来脱敏字段
	if s, ok := req.(Sanitizer); ok {
		return s.Sanitize()
	}

	// 1. 将 req 转为 JSON 字节
	b, err := json.Marshal(req)
	if err != nil {
		return "***(marshal_error)***" // 序列化失败时兜底脱敏
	}

	// 2. 解析为动态 map
	var data map[string]interface{}
	if err := json.Unmarshal(b, &data); err != nil {
		// 如果不是结构体/Map（如普通字符串），直接打码返回以策安全
		return "***(unsupported_type)***"
	}

	// 3. 递归脱敏处理
	maskSensitiveData(data, sensitizedMap)

	// 4. 重新转为 JSON 字符串用于日志打印
	safeBytes, _ := json.Marshal(data)
	return string(safeBytes)
}

// maskSensitiveData 递归遍历 Map 并根据名单打码
func maskSensitiveData(data map[string]interface{}, sensitizedMap map[string]bool) {
	for k, v := range data {
		// 如果值是嵌套的 map，递归处理
		if subMap, ok := v.(map[string]interface{}); ok {
			maskSensitiveData(subMap, sensitizedMap)
			continue
		}

		// 如果字段不在白名单中，进行脱敏（替换为 ***）
		if sensitizedMap[k] {
			data[k] = "***"
		}
	}
}
