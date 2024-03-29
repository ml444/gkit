package main

import (
	"bytes"
	_ "embed"
	"fmt"
	"regexp"
	"strings"
	"text/template"

	"google.golang.org/protobuf/compiler/protogen"
)

const release = "v1.0.0"

//go:embed field.tmpl
var fieldTemplate string

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

func generateFile(gen *protogen.Plugin, file *protogen.File, checkFuncs []checkFunc) *protogen.GeneratedFile {
	if len(file.Messages) == 0 {
		return nil
	}
	filename := file.GeneratedFilenamePrefix + "_field.pb.go"
	g := gen.NewGeneratedFile(filename, file.GoImportPath)
	g.P("// Code generated by protoc-gen-go-field. DO NOT EDIT.")
	g.P("// versions:")
	g.P(fmt.Sprintf("// - protoc-gen-go-field %s", release))
	g.P("// - protoc             ", protocVersion(gen))
	if file.Proto.GetOptions().GetDeprecated() {
		g.P("// ", file.Desc.Path(), " is a deprecated file.")
	} else {
		g.P("// source: ", file.Desc.Path())
	}
	g.P()
	g.P("package ", file.GoPackageName)
	err := genContent(file, g, checkFuncs)
	if err != nil {
		gen.Error(err)
	}
	return g
}

func genContent(file *protogen.File, g *protogen.GeneratedFile, checkFuncs []checkFunc) error {
	var fieldMap = make(map[string]string)
	for _, message := range file.Messages {
		if message.Desc.IsMapEntry() {
			continue
		}
		messageName := string(message.Desc.Name())
		if !chainCheck(messageName, checkFuncs...) {
			continue
		}

		for goName, pName := range getMessageFields(message) {
			if _, ok := fieldMap[goName]; ok {
				continue
			}
			fieldMap[goName] = pName
		}

	}

	tmpl, err := template.New("field").Parse(strings.TrimSpace(fieldTemplate))
	if err != nil {
		return err
	}
	if len(fieldMap) > 0 {
		buf := new(bytes.Buffer)
		e := tmpl.Execute(buf, fieldMap)
		if e != nil {
			panic(e.Error())
		}
		_, e = g.Write(buf.Bytes())
		if e != nil {
			panic(e.Error())
		}
	}
	return nil
}

func getMessageFields(message *protogen.Message) map[string]string {
	var sMap = make(map[string]string)
	for _, field := range message.Fields {
		sMap[field.GoName] = string(field.Desc.Name())
	}
	return sMap
}

type checkFunc func(messageName string) (ok bool)

func chainCheck(s string, fnList ...checkFunc) bool {
	for _, fn := range fnList {
		if !fn(s) {
			return false
		}
	}
	return true
}

func checkIncludePrefix(prefix string) checkFunc {
	return func(s string) bool {
		return s == prefix || len(prefix) == 0 || len(s) > len(prefix) && s[:len(prefix)] == prefix
	}
}

func checkExcludePrefix(prefix string) checkFunc {
	return func(s string) bool {
		return s != prefix && !(len(prefix) > 0 && len(s) > len(prefix) && s[:len(prefix)] == prefix)
	}
}

func checkIncludeSuffix(suffix string) checkFunc {
	return func(s string) bool {
		return s == suffix || len(suffix) == 0 || len(s) > len(suffix) && s[len(s)-len(suffix):] == suffix
	}
}

func checkExcludeSuffix(suffix string) checkFunc {
	return func(s string) bool {
		return s != suffix && !(len(suffix) > 0 && len(s) > len(suffix) && s[len(s)-len(suffix):] == suffix)
	}
}

func checkIncludePattern(pattern string) checkFunc {
	regPattern := regexp.MustCompile(pattern)
	return func(s string) bool {
		return regPattern.MatchString(s)
	}
}

func checkExcludePattern(pattern string) checkFunc {
	regPattern := regexp.MustCompile(pattern)
	return func(s string) bool {
		return !regPattern.MatchString(s)
	}
}
