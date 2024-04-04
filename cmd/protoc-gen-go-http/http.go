package main

import (
	"fmt"
	"net/http"
	"os"
	"strings"

	"google.golang.org/protobuf/reflect/protoreflect"

	"google.golang.org/genproto/googleapis/api/annotations"
	"google.golang.org/protobuf/compiler/protogen"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/descriptorpb"
)

const release = "v1.1.0"
const deprecationComment = "// Deprecated: Do not use."

const (
	httpPackage       = protogen.GoImportPath("net/http")
	contextPackage    = protogen.GoImportPath("context")
	pluckPackage      = protogen.GoImportPath("github.com/ml444/gkit/cmd/protoc-gen-go-http/pluck")
	middlewarePackage = protogen.GoImportPath("github.com/ml444/gkit/middleware")
	transportPackage  = protogen.GoImportPath("github.com/ml444/gkit/transport")
	httpxPackage      = protogen.GoImportPath("github.com/ml444/gkit/transport/httpx")
)

var methodSets = make(map[string]int)

// var hasPluck bool

func generateFile(gen *protogen.Plugin, file *protogen.File, omitempty bool, omitemptyPrefix string) *protogen.GeneratedFile {
	if len(file.Services) == 0 || (omitempty && !hasHTTPRule(file.Services)) {
		return nil
	}
	filename := file.GeneratedFilenamePrefix + "_http.pb.go"
	g := gen.NewGeneratedFile(filename, file.GoImportPath)
	g.P("// Code generated by protoc-gen-go-http. DO NOT EDIT.")
	g.P("// versions:")
	g.P(fmt.Sprintf("// - protoc-gen-go-http %s", release))
	g.P("// - protoc             ", protocVersion(gen))
	if file.Proto.GetOptions().GetDeprecated() {
		g.P("// ", file.Desc.Path(), " is a deprecated file.")
	} else {
		g.P("// source: ", file.Desc.Path())
	}
	g.P()
	g.P("package ", file.GoPackageName)
	g.P()
	generateFileContent(gen, file, g, omitempty, omitemptyPrefix)
	return g
}

func generateFileContent(gen *protogen.Plugin, file *protogen.File, g *protogen.GeneratedFile, omitempty bool, omitemptyPrefix string) {
	if len(file.Services) == 0 {
		return
	}
	g.P("var _ = new(", httpPackage.Ident("Request"), ")")
	g.P("var _ = new(", contextPackage.Ident("Context"), ")")
	g.P("var _  = make([]", middlewarePackage.Ident("Middleware"), ", 0)")
	g.P("var _ ", transportPackage.Ident("Server"), " = new(", httpxPackage.Ident("Server"), ")")
	g.P("var _ = ", pluckPackage.Ident("DisablePluckHeader"))
	g.P()

	for _, service := range file.Services {
		genService(gen, file, g, service, omitempty, omitemptyPrefix)
	}
}

func genService(_ *protogen.Plugin, file *protogen.File, g *protogen.GeneratedFile, service *protogen.Service, omitempty bool, omitemptyPrefix string) {
	if service.Desc.Options().(*descriptorpb.ServiceOptions).GetDeprecated() {
		g.P("//")
		g.P(deprecationComment)
	}
	// HTTP Server.
	sd := &serviceCtx{
		ServiceType: service.GoName,
		ServiceName: string(service.Desc.FullName()),
		Metadata:    file.Desc.Path(),
	}
	for _, method := range service.Methods {
		if method.Desc.IsStreamingClient() || method.Desc.IsStreamingServer() {
			continue
		}
		rule, ok := proto.GetExtension(method.Desc.Options(), annotations.E_Http).(*annotations.HttpRule)
		if rule != nil && ok {
			for _, bind := range rule.AdditionalBindings {
				sd.Methods = append(sd.Methods, buildHTTPRule(g, service, method, bind, omitemptyPrefix))
			}
			sd.Methods = append(sd.Methods, buildHTTPRule(g, service, method, rule, omitemptyPrefix))
		} else if !omitempty {
			path := fmt.Sprintf("%s/%s/%s", omitemptyPrefix, service.Desc.FullName(), method.Desc.Name())
			sd.Methods = append(sd.Methods, buildMethodDesc(g, method, http.MethodPost, path))
		}
	}
	if len(sd.Methods) != 0 {
		g.P(sd.execute())
	}
}

func hasHTTPRule(services []*protogen.Service) bool {
	for _, service := range services {
		for _, method := range service.Methods {
			if method.Desc.IsStreamingClient() || method.Desc.IsStreamingServer() {
				continue
			}
			rule, ok := proto.GetExtension(method.Desc.Options(), annotations.E_Http).(*annotations.HttpRule)
			if rule != nil && ok {
				return true
			}
		}
	}
	return false
}

func buildHTTPRule(g *protogen.GeneratedFile, service *protogen.Service, m *protogen.Method, rule *annotations.HttpRule, omitemptyPrefix string) *methodCtx {
	var (
		path         string
		method       string
		body         string
		responseBody string
	)

	switch pattern := rule.Pattern.(type) {
	case *annotations.HttpRule_Get:
		path = pattern.Get
		method = http.MethodGet
	case *annotations.HttpRule_Put:
		path = pattern.Put
		method = http.MethodPut
	case *annotations.HttpRule_Post:
		path = pattern.Post
		method = http.MethodPost
	case *annotations.HttpRule_Delete:
		path = pattern.Delete
		method = http.MethodDelete
	case *annotations.HttpRule_Patch:
		path = pattern.Patch
		method = http.MethodPatch
	case *annotations.HttpRule_Custom:
		path = pattern.Custom.Path
		method = pattern.Custom.Kind
	}
	if method == "" {
		method = http.MethodPost
	}
	if path == "" {
		path = fmt.Sprintf("%s/%s/%s", omitemptyPrefix, service.Desc.FullName(), m.Desc.Name())
	}
	body = rule.Body
	responseBody = rule.ResponseBody
	md := buildMethodDesc(g, m, method, path)
	if method == http.MethodGet || method == http.MethodDelete {
		if body != "" {
			_, _ = fmt.Fprintf(os.Stderr, "\u001B[31mWARN\u001B[m: %s %s body should not be declared.\n", method, path)
		}
	} else {
		if body == "" {
			_, _ = fmt.Fprintf(os.Stderr, "\u001B[31mWARN\u001B[m: %s %s does not declare a body.\n", method, path)
		}
	}
	if body == "*" {
		md.HasBody = true
		md.Body = ""
	} else if body != "" {
		md.HasBody = true
		bodyField := camelCaseVars(body)
		md.Body = "." + bodyField
		md.BodyFieldIsBytes(m.Input, body)
		// md.ReqBodyToField = bodyField
	} else {
		md.HasBody = false
	}
	if responseBody == "*" {
		md.ResponseBody = ""
	} else if responseBody != "" {
		rspBodyField := camelCaseVars(responseBody)
		md.ResponseBody = "." + rspBodyField
		// md.FieldToRspBody = rspBodyField
	}
	md.ParsePluck(m)
	return md
}

func buildMethodDesc(g *protogen.GeneratedFile, m *protogen.Method, method, path string) *methodCtx {
	defer func() { methodSets[m.GoName]++ }()

	vars := buildPathVars(path)

	for v, s := range vars {
		fields := m.Input.Desc.Fields()

		if s != nil {
			path = replacePath(v, *s, path)
		}
		for _, field := range strings.Split(v, ".") {
			if strings.TrimSpace(field) == "" {
				continue
			}
			if strings.Contains(field, ":") {
				field = strings.Split(field, ":")[0]
			}
			fd := fields.ByName(protoreflect.Name(field))
			if fd == nil {
				fmt.Fprintf(os.Stderr, "\u001B[31mERROR\u001B[m: The corresponding field '%s' declaration in message could not be found in '%s'\n", v, path)
				os.Exit(2)
			}
			if fd.IsMap() {
				fmt.Fprintf(os.Stderr, "\u001B[31mWARN\u001B[m: The field in path:'%s' shouldn't be a map.\n", v)
			} else if fd.IsList() {
				fmt.Fprintf(os.Stderr, "\u001B[31mWARN\u001B[m: The field in path:'%s' shouldn't be a list.\n", v)
			} else if fd.Kind() == protoreflect.MessageKind || fd.Kind() == protoreflect.GroupKind {
				fields = fd.Message().Fields()
			}
		}
	}
	comment := m.Comments.Leading.String() + m.Comments.Trailing.String()
	if comment != "" {
		comment = "// " + m.GoName + strings.TrimPrefix(strings.TrimSuffix(comment, "\n"), "//")
	}
	return &methodCtx{
		Name:         m.GoName,
		OriginalName: string(m.Desc.Name()),
		Num:          methodSets[m.GoName],
		Input:        g.QualifiedGoIdent(m.Input.GoIdent),
		Output:       g.QualifiedGoIdent(m.Output.GoIdent),
		Comment:      comment,
		Path:         path,
		Method:       method,
		HasVars:      len(vars) > 0,
	}
}

func protocVersion(gen *protogen.Plugin) string {
	v := gen.Request.GetCompilerVersion()
	if v == nil {
		return "(unknown)"
	}
	var suffix string
	if s := v.GetSuffix(); s != "" {
		suffix = "-" + s
	}
	return fmt.Sprintf("v%d.%d.%d%s", v.GetMajor(), v.GetMinor(), v.GetPatch(), suffix)
}
