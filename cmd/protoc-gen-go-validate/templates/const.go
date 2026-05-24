package templates

const ConstTpl = `{{ $f := .Field }}{{ $r := .Rules }}
	{{- if $r.Const }}
		if {{ .GetAccessor }} != {{ lit $r.GetConst }} {
			{{- if isEnum $f }}
			err := {{GetAliasName}}ValidationError(
				{{.ErrCode}}, 
				"[{{ $f.GoName }}] value must equal {{ enumVal $f $r.GetConst }}",
				nil,
			)
			{{- else }}
			err := {{GetAliasName}}ValidationError(
				{{.ErrCode}}, 
				fmt.Sprintf("[{{ $f.GoName }}] value must equal %s", {{ lit $r.GetConst }}),
				nil,
			)
			{{- end }}
			if !all { return err }
			errors = append(errors, err)
		}
	{{- end }}
`
