package main

import (
	"flag"
	"strings"

	"google.golang.org/protobuf/compiler/protogen"
	"google.golang.org/protobuf/types/pluginpb"
)

var needGenFieldFunc = flag.String("field_func", "CreatedAt,UpdatedAt,DeletedAt", "Specifies which fields in the model's structure need to generate field functions")
var NeedGenerateFunctionFields map[string]bool

func main() {
	flag.Parse()
	if needGenFieldFunc != nil && *needGenFieldFunc != "" {
		NeedGenerateFunctionFields = make(map[string]bool)
		for _, v := range strings.Split(*needGenFieldFunc, ",") {
			NeedGenerateFunctionFields[v] = true
		}
	}

	protogen.Options{
		ParamFunc: flag.CommandLine.Set,
	}.Run(func(gen *protogen.Plugin) error {
		gen.SupportedFeatures = uint64(pluginpb.CodeGeneratorResponse_FEATURE_PROTO3_OPTIONAL)
		for _, f := range gen.Files {
			if !f.Generate {
				continue
			}
			generateFile(gen, f)
		}
		return nil
	})
}
