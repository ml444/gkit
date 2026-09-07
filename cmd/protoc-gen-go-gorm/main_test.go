package main

import (
	"testing"

	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/descriptorpb"
	"google.golang.org/protobuf/types/pluginpb"
)

func TestSourceRelativeShorthand(t *testing.T) {
	file := &descriptorpb.FileDescriptorProto{
		Name:    proto.String("api/v1/pool.proto"),
		Package: proto.String("pool"),
		Syntax:  proto.String("proto3"),
		Options: &descriptorpb.FileOptions{GoPackage: proto.String("example.com/project/models/pool;pool")},
	}
	sourceRelative := false
	gen, err := protogenOptions(&sourceRelative).New(&pluginpb.CodeGeneratorRequest{
		Parameter:      proto.String("source_relative"),
		ProtoFile:      []*descriptorpb.FileDescriptorProto{file},
		FileToGenerate: []string{file.GetName()},
	})
	if err != nil {
		t.Fatalf("create protogen plugin: %v", err)
	}
	if !sourceRelative {
		t.Fatal("source_relative shorthand was not enabled")
	}

	applySourceRelativePaths(gen)

	if got, want := gen.Files[0].GeneratedFilenamePrefix, "api/v1/pool"; got != want {
		t.Fatalf("generated filename prefix = %q, want %q", got, want)
	}
}
