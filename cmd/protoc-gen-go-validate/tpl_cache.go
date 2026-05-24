package main

import (
	"fmt"
	"strings"
	"sync"
	"text/template"

	"github.com/ml444/gkit/cmd/protoc-gen-go-validate/funcs"
)

var (
	subTemplates     *template.Template
	subTemplatesOnce sync.Once
	subTemplatesErr  error
)

func initSubTemplates() {
	subTemplates = template.New("file")
	Register(subTemplates)
}

func getFileTemplate() (*template.Template, error) {
	subTemplatesOnce.Do(func() {
		initSubTemplates()
		if subTemplatesErr != nil {
			return
		}
	})
	if subTemplatesErr != nil {
		return nil, subTemplatesErr
	}

	tmpl, err := subTemplates.Clone()
	if err != nil {
		return nil, fmt.Errorf("clone template: %w", err)
	}

	funcMap := copyFuncMap()
	funcMap["render"] = funcs.Render(tmpl)
	tmpl = tmpl.Funcs(funcMap)

	_, err = tmpl.New("validate").Parse(strings.TrimSpace(validateTemplate))
	if err != nil {
		return nil, fmt.Errorf("parse validate template: %w", err)
	}
	return tmpl, nil
}

func copyFuncMap() map[string]interface{} {
	m := make(map[string]interface{}, len(funcs.FuncMap)+1)
	for k, v := range funcs.FuncMap {
		m[k] = v
	}
	return m
}
