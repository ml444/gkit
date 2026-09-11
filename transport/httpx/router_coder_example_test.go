package httpx_test

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"

	"github.com/ml444/gkit/transport/httpx"
	jsoncodec "github.com/ml444/gkit/transport/httpx/coder/json"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/descriptorpb"
)

func ExampleNewRouterCoder() {
	opts := jsoncodec.DefaultOptions()
	opts.Marshal.UseProtoNames = true
	opts.Marshal.EmitUnpopulated = false
	rc, err := httpx.NewRouterCoder(
		httpx.WithJSONCoder(jsoncodec.NewCoder(opts)),
		httpx.WithBindQuery(func(r *http.Request, target interface{}) error {
			field := target.(*descriptorpb.FieldDescriptorProto)
			field.JsonName = proto.String(strings.ToUpper(r.URL.Query().Get("name")))
			return nil
		}),
	)
	if err != nil {
		panic(err)
	}
	srv := httpx.NewServer(httpx.RouterCoder(rc))
	srv.GetRouter().GET("/field", func(w http.ResponseWriter, r *http.Request) {
		ctx := httpx.NewCtx(w, r)
		var field descriptorpb.FieldDescriptorProto
		if err := ctx.BindQuery(&field); err != nil {
			ctx.ReturnError(err)
			return
		}
		ctx.Result(http.StatusOK, &field)
	})
	w := httptest.NewRecorder()
	srv.ServeHTTP(w, httptest.NewRequest(http.MethodGet, "/field?name=display", nil))
	var fields map[string]string
	if err := json.Unmarshal(w.Body.Bytes(), &fields); err != nil {
		panic(err)
	}
	fmt.Println(w.Code)
	fmt.Println(fields["json_name"])
	// Output:
	// 200
	// DISPLAY
}
