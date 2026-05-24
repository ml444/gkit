package templates

const InTpl = `{{ $f := .Field }}{{ $r := .Rules }}
	{{- if $r.In }}
		if _, ok := {{ lookup $f "InLookup" }}[{{ if isBytes $f.Desc }}string({{ .GetAccessor }}){{ else }}{{ .GetAccessor }}{{ end }}]; !ok {
			{{- if isEnum $f }}
			err := {{GetAliasName}}ValidationError(
				{{.ErrCode}}, 
				"[{{ $f.GoName }}] value must be in {{ enumList $f $r.In }}",
				nil,
			)
			{{- else }}
			err := {{GetAliasName}}ValidationError(
				{{.ErrCode}}, 
				"[{{ $f.GoName }}] value must be in {{ $r.In }}",
				nil,
			)
			{{- end }}
			if !all { return err }
			errors = append(errors, err)
		}
	{{- else if $r.NotIn }}
		if _, ok := {{ lookup $f "NotInLookup" }}[{{ if isBytes $f.Desc }}string({{ .GetAccessor }}){{ else }}{{ .GetAccessor }}{{ end }}]; ok {
			{{- if isEnum $f }}
			err := {{GetAliasName}}ValidationError(
				{{.ErrCode}}, 
				"[{{ $f.GoName }}] value must not be in {{ enumList $f $r.NotIn }}",
				nil,
			)
			{{- else }}
			err := {{GetAliasName}}ValidationError(
				{{.ErrCode}}, 
				"[{{ $f.GoName }}] value must not be in {{ $r.NotIn }}",
				nil,
			)
			{{- end }}
			if !all { return err }
			errors = append(errors, err)
		}
	{{- end }}
`
