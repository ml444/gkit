package main

import (
	"text/template"

	"github.com/ml444/gkit/cmd/protoc-gen-go-validate/funcs"
	"github.com/ml444/gkit/cmd/protoc-gen-go-validate/templates"
)

func Register(tpl *template.Template) {
	funcMap := funcs.FuncMap
	funcMap["render"] = funcs.Render(tpl)
	tpl.Funcs(funcMap)

	template.Must(tpl.New("required").Parse(templates.RequiredTpl))
	template.Must(tpl.New("timestamp").Parse(templates.TimestampTpl))
	template.Must(tpl.New("duration").Parse(templates.DurationTpl))
	template.Must(tpl.New("message").Parse(templates.MessageTpl))
	//template.Must(tpl.New("msg").Parse(templates.MsgTpl))
	template.Must(tpl.New("const").Parse(templates.ConstTpl))
	template.Must(tpl.New("ltgt").Parse(templates.LtGtTpl))
	template.Must(tpl.New("in").Parse(templates.InTpl))

	template.Must(tpl.New("none").Parse(templates.NoneTpl))
	template.Must(tpl.New("float").Parse(templates.NumTpl))
	template.Must(tpl.New("double").Parse(templates.NumTpl))
	template.Must(tpl.New("int32").Parse(templates.NumTpl))
	template.Must(tpl.New("int64").Parse(templates.NumTpl))
	template.Must(tpl.New("uint32").Parse(templates.NumTpl))
	template.Must(tpl.New("uint64").Parse(templates.NumTpl))
	template.Must(tpl.New("sint32").Parse(templates.NumTpl))
	template.Must(tpl.New("sint64").Parse(templates.NumTpl))
	template.Must(tpl.New("fixed32").Parse(templates.NumTpl))
	template.Must(tpl.New("fixed64").Parse(templates.NumTpl))
	template.Must(tpl.New("sfixed32").Parse(templates.NumTpl))
	template.Must(tpl.New("sfixed64").Parse(templates.NumTpl))

	template.Must(tpl.New("bool").Parse(templates.ConstTpl))
	template.Must(tpl.New("string").Parse(templates.StrTpl))
	template.Must(tpl.New("bytes").Parse(templates.BytesTpl))

	template.Must(tpl.New("email").Parse(templates.EmailTpl))
	template.Must(tpl.New("hostname").Parse(templates.HostTpl))
	template.Must(tpl.New("address").Parse(templates.HostTpl))
	template.Must(tpl.New("uuid").Parse(templates.UuidTpl))

	template.Must(tpl.New("enum").Parse(templates.EnumTpl))
	template.Must(tpl.New("repeated").Parse(templates.RepTpl))
	template.Must(tpl.New("map").Parse(templates.MapTpl))

	template.Must(tpl.New("any").Parse(templates.AnyTpl))
	template.Must(tpl.New("timestampcmp").Parse(templates.TimestampcmpTpl))
	template.Must(tpl.New("durationcmp").Parse(templates.DurationcmpTpl))

	template.Must(tpl.New("wrapper").Parse(templates.WrapperTpl))
}
