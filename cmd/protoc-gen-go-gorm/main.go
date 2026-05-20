package main

import (
	"flag"
	"strings"

	"google.golang.org/protobuf/compiler/protogen"
	"google.golang.org/protobuf/types/pluginpb"
)

var needGenFieldFunc = flag.String(
	"field_func",
	"",
	"Comma-separated TModel field names to generate getters for (optional; proto getters are preferred for dbx)",
)

func parseFieldFuncFlag() map[string]bool {
	if needGenFieldFunc == nil || *needGenFieldFunc == "" {
		return nil
	}
	out := make(map[string]bool)
	for _, v := range strings.Split(*needGenFieldFunc, ",") {
		v = strings.TrimSpace(v)
		if v != "" {
			out[v] = true
		}
	}
	return out
}

func main() {
	flag.Parse()
	fieldFuncs := parseFieldFuncFlag()

	protogen.Options{
		ParamFunc: flag.CommandLine.Set,
	}.Run(func(gen *protogen.Plugin) error {
		gen.SupportedFeatures = uint64(pluginpb.CodeGeneratorResponse_FEATURE_PROTO3_OPTIONAL)
		for _, f := range gen.Files {
			if !f.Generate {
				continue
			}
			generateFile(gen, f, fieldFuncs)
		}
		return nil
	})
}
