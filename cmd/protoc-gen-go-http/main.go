package main

import (
	"flag"
	"fmt"

	"google.golang.org/protobuf/compiler/protogen"
	"google.golang.org/protobuf/types/pluginpb"
)

var (
	showVersion     = flag.Bool("version", false, "print the version and exit")
	omitempty       = flag.Bool("omitempty", true, "omit if google.api is empty")
	omitemptyPrefix = flag.String("omitempty_prefix", "", "prefix for default POST paths when omitempty=false")
	module          = flag.String("module", "", "base Go module path for generated imports (default: github.com/ml444/gkit)")
	clientMode      = flag.String("client", "full", "client generation mode: full or none")
	warnings        = flag.String("warnings", "warn", "warning level: warn, off, or error")
)

func main() {
	flag.Parse()
	if *showVersion {
		fmt.Printf("protoc-gen-go-http %v\n", release)
		return
	}
	cfg := newPluginConfig(*omitempty, *omitemptyPrefix, *module, *clientMode, *warnings)
	protogen.Options{
		ParamFunc: flag.CommandLine.Set,
	}.Run(func(gen *protogen.Plugin) error {
		gen.SupportedFeatures = uint64(pluginpb.CodeGeneratorResponse_FEATURE_PROTO3_OPTIONAL)
		for _, f := range gen.Files {
			if !f.Generate {
				continue
			}
			if err := generateFile(gen, f, cfg); err != nil {
				return err
			}
		}
		return nil
	})
}
