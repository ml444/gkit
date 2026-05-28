package main

import (
	"bytes"
	_ "embed"
	"fmt"
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
	ServiceType    string
	ServiceName    string
	Metadata       string
	Methods        []*methodCtx
	MethodSets     map[string]*methodCtx
	BindingCount   map[string]int
	GenerateClient bool
	Cfg            pluginConfig
}

func (s *serviceCtx) execute() string {
	s.MethodSets = make(map[string]*methodCtx)
	s.BindingCount = make(map[string]int)
	for _, m := range s.Methods {
		s.MethodSets[m.Name] = m
		s.BindingCount[m.Name]++
	}
	for _, m := range s.Methods {
		m.BindingsForRPC = s.BindingCount[m.Name]
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

func (s *serviceCtx) hasRawHandlers() bool {
	for _, m := range s.Methods {
		if m.HasRawResponse {
			return true
		}
	}
	return false
}

type methodCtx struct {
	pluckDesc
	Name         string
	OriginalName string
	Num          int
	Input        string
	Output       string
	Comment      string
	Path         string
	Method       string
	Body         string
	ResponseBody string
	HasVars      bool
	HasBody      bool
	BodyIsBytes  bool
	BindingsForRPC int
}

type pluckDesc struct {
	ReqHeadersToField string
	ReqBodyToField    string
	RspHeadersField   string
	HasPluck          bool
	HasRawResponse    bool
	DefaultContentType string
	ReqHeaders        map[string]string
	RspHeaders        map[string]string
}

func (m *methodCtx) ClientName() string {
	if m.BindingsForRPC <= 1 {
		return m.Name
	}
	if m.Num == 0 {
		return m.Name
	}
	return fmt.Sprintf("%s_%d", m.Name, m.Num)
}

func (m *methodCtx) setResponseBodyField(field string) {
	if field == "" {
		return
	}
	m.ResponseBody = "." + field
	m.HasRawResponse = true
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
}

func isExistField(desc protoreflect.MessageDescriptor, field string) bool {
	fields := desc.Fields()
	parts := strings.Split(field, ".")
	for i, f := range parts {
		fd := fields.ByName(protoreflect.Name(f))
		if fd == nil {
			return false
		}
		if i == len(parts)-1 {
			return true
		}
		if fd.Kind() != protoreflect.MessageKind {
			return false
		}
		fields = fd.Message().Fields()
	}
	return false
}

func (md *methodCtx) ParsePluck(method *protogen.Method) error {
	pluckReqOpt, ok := proto.GetExtension(method.Desc.Options(), pluck.E_Request).(*pluck.PluckRequest)
	if ok && pluckReqOpt != nil {
		if pluckReqOpt.DefaultHeaders != nil {
			md.HasPluck = true
			md.ReqHeaders = pluckFields(pluckReqOpt.DefaultHeaders)
		}
		if pluckReqOpt.HeadersTo != "" {
			md.HasPluck = true
			md.ReqHeadersToField = camelCase(pluckReqOpt.HeadersTo)
			if !isExistField(method.Input.Desc, pluckReqOpt.HeadersTo) {
				return fmt.Errorf("%s: pluck.request.headers_to field %q not found in request message", method.Desc.FullName(), pluckReqOpt.HeadersTo)
			}
		}
		if pluckReqOpt.BodyTo != "" {
			md.HasPluck = true
			md.ReqBodyToField = camelCase(pluckReqOpt.BodyTo)
			md.BodyFieldIsBytes(method.Input, pluckReqOpt.BodyTo)
			if !isExistField(method.Input.Desc, pluckReqOpt.BodyTo) {
				return fmt.Errorf("%s: pluck.request.body_to field %q not found in request message", method.Desc.FullName(), pluckReqOpt.BodyTo)
			}
		}
	}
	pluckRspOpt, ok := proto.GetExtension(method.Desc.Options(), pluck.E_Response).(*pluck.PluckResponse)
	if ok && pluckRspOpt != nil {
		if pluckRspOpt.DefaultHeaders != nil {
			md.HasPluck = true
			md.RspHeaders = pluckFields(pluckRspOpt.DefaultHeaders)
			if ct, ok := md.RspHeaders["Content-Type"]; ok && md.DefaultContentType == "" {
				md.DefaultContentType = ct
			}
		}
		if pluckRspOpt.HeadersFrom != "" {
			md.HasPluck = true
			md.RspHeadersField = camelCase(pluckRspOpt.HeadersFrom)
			if !isExistField(method.Output.Desc, pluckRspOpt.HeadersFrom) {
				return fmt.Errorf("%s: pluck.response.headers_from field %q not found in response message", method.Desc.FullName(), pluckRspOpt.HeadersFrom)
			}
		}
		if pluckRspOpt.BodyFrom != "" {
			md.HasPluck = true
			md.setResponseBodyField(camelCase(pluckRspOpt.BodyFrom))
			if !isExistField(method.Output.Desc, pluckRspOpt.BodyFrom) {
				return fmt.Errorf("%s: pluck.response.body_from field %q not found in response message", method.Desc.FullName(), pluckRspOpt.BodyFrom)
			}
		}
	}
	return nil
}
