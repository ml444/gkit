package main

import (
	"fmt"
	"os"
	"strings"

	"google.golang.org/protobuf/compiler/protogen"
)

const defaultModule = "github.com/ml444/gkit"
const release = "v1.2.0"
const deprecationComment = "// Deprecated: Do not use."

type pluginConfig struct {
	omitempty       bool
	omitemptyPrefix string
	module          string
	clientMode      string // full, none
	warnings        string // warn, off, error
}

func newPluginConfig(omitempty bool, omitemptyPrefix, module, clientMode, warnings string) pluginConfig {
	cfg := pluginConfig{
		omitempty:       omitempty,
		omitemptyPrefix: omitemptyPrefix,
		module:          strings.TrimSpace(module),
		clientMode:      clientMode,
		warnings:        warnings,
	}
	if cfg.module == "" {
		cfg.module = defaultModule
	}
	if cfg.clientMode == "" {
		cfg.clientMode = "full"
	}
	if cfg.warnings == "" {
		cfg.warnings = "warn"
	}
	return cfg
}

func (c pluginConfig) httpxPackage() protogen.GoImportPath {
	return protogen.GoImportPath(c.module + "/transport/httpx")
}

func (c pluginConfig) middlewarePackage() protogen.GoImportPath {
	return protogen.GoImportPath(c.module + "/middleware")
}

func (c pluginConfig) pluckPackage() protogen.GoImportPath {
	return protogen.GoImportPath(c.module + "/cmd/protoc-gen-go-http/pluck")
}

func (c pluginConfig) responsePackage() protogen.GoImportPath {
	return protogen.GoImportPath(c.module + "/middleware/response")
}

func (c pluginConfig) generateClient() bool {
	return c.clientMode != "none"
}

func (c pluginConfig) warn(format string, args ...interface{}) error {
	if c.warnings == "off" {
		return nil
	}
	msg := fmt.Sprintf(format, args...)
	_, _ = fmt.Fprintf(os.Stderr, "\u001B[33mWARN\u001B[m: %s\n", msg)
	if c.warnings == "error" {
		return fmt.Errorf("%s", msg)
	}
	return nil
}
