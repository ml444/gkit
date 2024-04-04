package main

import (
	"bytes"
	_ "embed"
	"os"
	"strings"
	"text/template"

	"google.golang.org/protobuf/compiler/protogen"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/reflect/protoreflect"

	"github.com/ml444/gkit/cmd/protoc-gen-go-http/pluck"
)

//go:embed http.tmpl
var httpTemplate string

type serviceCtx struct {
	ServiceType string // Greeter
	ServiceName string // helloworld.Greeter
	Metadata    string // api/helloworld/helloworld.proto
	Methods     []*methodCtx
	MethodSets  map[string]*methodCtx
}

func (s *serviceCtx) execute() string {
	s.MethodSets = make(map[string]*methodCtx)
	for _, m := range s.Methods {
		s.MethodSets[m.Name] = m
	}
	buf := new(bytes.Buffer)
	tmpl, err := template.New("http").Parse(strings.TrimSpace(httpTemplate))
	if err != nil {
		panic(err)
	}
	if err := tmpl.Execute(buf, s); err != nil {
		panic(err)
	}
	return strings.Trim(buf.String(), "\r\n")
}

type methodCtx struct {
	pluckDesc
	// method
	Name         string
	OriginalName string // The parsed original name
	Num          int
	Input        string
	Output       string
	Comment      string
	// http_rule
	Path         string
	Method       string
	Body         string
	ResponseBody string
	HasVars      bool
	HasBody      bool
	BodyIsBytes  bool
}

type pluckDesc struct {
	ReqHeadersToField string
	ReqBodyToField    string

	FieldToRspHeaders string
	FieldToRspBody    string

	HasPluck bool

	ReqHeaders map[string]string
	RspHeaders map[string]string
}

func (md *methodCtx) BodyFieldIsBytes(m *protogen.Message, field string) {
	fields := m.Desc.Fields()
	for _, f := range strings.Split(field, ".") {
		fd := fields.ByName(protoreflect.Name(f))
		if fd == nil {
			return
		}
		if fd.Kind() == protoreflect.BytesKind {
			md.BodyIsBytes = true
			return
		}
		if fd.Kind() == protoreflect.MessageKind {
			fields = fd.Message().Fields()
		}
	}
	return
}

func (md *methodCtx) isExistField(m *protogen.Message, field string) bool {
	fields := m.Desc.Fields()
	for _, f := range strings.Split(field, ".") {
		fd := fields.ByName(protoreflect.Name(f))
		if fd == nil {
			break
		}
		if fd.Name() == protoreflect.Name(f) {
			return true
		}
		if fd.Kind() == protoreflect.MessageKind {
			fields = fd.Message().Fields()
		}
	}
	println(string(m.Desc.FullName()) + " not found the field: " + field)
	return false
}

func (md *methodCtx) ParsePluck(method *protogen.Method) {
	pluckReqOpt, ok := proto.GetExtension(method.Desc.Options(), pluck.E_Request).(*pluck.PluckRequest)
	if ok && pluckReqOpt != nil {
		if pluckReqOpt.DefaultHeaders != nil {
			md.HasPluck = true
			md.ReqHeaders = pluckFields(pluckReqOpt.DefaultHeaders)
		}
		if pluckReqOpt.HeadersTo != "" {
			md.HasPluck = true
			md.ReqHeadersToField = camelCase(pluckReqOpt.HeadersTo)
			if !md.isExistField(method.Input, pluckReqOpt.HeadersTo) {
				os.Exit(2)
			}
		}
		if pluckReqOpt.BodyTo != "" {
			md.HasPluck = true
			md.ReqBodyToField = camelCase(pluckReqOpt.BodyTo)
			md.BodyFieldIsBytes(method.Input, pluckReqOpt.BodyTo)
			if !md.isExistField(method.Input, pluckReqOpt.BodyTo) {
				os.Exit(2)
			}
		}
	}
	pluckRspOpt, ok := proto.GetExtension(method.Desc.Options(), pluck.E_Response).(*pluck.PluckResponse)
	if ok && pluckRspOpt != nil {
		if pluckRspOpt.DefaultHeaders != nil {
			md.HasPluck = true
			md.RspHeaders = pluckFields(pluckRspOpt.DefaultHeaders)
		}
		if pluckRspOpt.HeadersFrom != "" {
			md.HasPluck = true
			md.FieldToRspHeaders = camelCase(pluckRspOpt.HeadersFrom)
			if !md.isExistField(method.Output, pluckRspOpt.HeadersFrom) {
				os.Exit(2)
			}
		}
		if pluckRspOpt.BodyFrom != "" {
			md.HasPluck = true
			md.FieldToRspBody = camelCase(pluckRspOpt.BodyFrom)
			if !md.isExistField(method.Output, pluckRspOpt.BodyFrom) {
				os.Exit(2)
			}
		}
	}
}
