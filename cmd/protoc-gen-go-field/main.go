package main

import (
	"flag"

	"google.golang.org/protobuf/compiler/protogen"
	"google.golang.org/protobuf/types/pluginpb"
)

var (
	includePrefix  = flag.String("include_prefix", "", "include if message name has this prefix")
	excludePrefix  = flag.String("exclude_prefix", "", "exclude if message name has this prefix")
	includeSuffix  = flag.String("include_suffix", "", "include if message name has this suffix")
	excludeSuffix  = flag.String("exclude_suffix", "", "exclude if message name has this suffix")
	includePattern = flag.String("include_pattern", "", "include if message name matches this pattern")
	excludePattern = flag.String("exclude_pattern", "", "exclude if message name matches this pattern")
)

func main() {
	flag.Parse()

	protogen.Options{
		ParamFunc: flag.CommandLine.Set,
	}.Run(func(gen *protogen.Plugin) error {
		gen.SupportedFeatures = uint64(pluginpb.CodeGeneratorResponse_FEATURE_PROTO3_OPTIONAL)
		for _, f := range gen.Files {
			if !f.Generate {
				continue
			}
			var checkFuncs []checkFunc
			if includePrefix != nil && *includePrefix != "" {
				checkFuncs = append(checkFuncs, checkIncludePrefix(*includePrefix))
			}
			if excludePrefix != nil && *excludePrefix != "" {
				checkFuncs = append(checkFuncs, checkExcludePrefix(*excludePrefix))
			}
			if includeSuffix != nil && *includeSuffix != "" {
				checkFuncs = append(checkFuncs, checkIncludeSuffix(*includeSuffix))
			}
			if excludeSuffix != nil && *excludeSuffix != "" {
				checkFuncs = append(checkFuncs, checkExcludeSuffix(*excludeSuffix))
			}
			if includePattern != nil && *includePattern != "" {
				checkFuncs = append(checkFuncs, checkIncludePattern(*includePattern))
			}
			if excludePattern != nil && *excludePattern != "" {
				checkFuncs = append(checkFuncs, checkExcludePattern(*excludePattern))
			}
			generateFile(gen, f, checkFuncs)
		}
		return nil
	})
}
