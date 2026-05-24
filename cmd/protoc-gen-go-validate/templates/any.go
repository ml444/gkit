package templates

const AnyTpl = `{{ $f := .Field }}{{ $r := .Rules }}
	{{- template "required" . }}

	if a := {{ .GetAccessor }}; a != nil {
		{{- if $r.In }}
			if _, ok := {{ lookup $f "InLookup" }}[a.GetTypeUrl()]; !ok {
				err := {{GetAliasName}}ValidationError(
					{{.ErrCode}}, 
					"[{{ $f.GoName }}] type URL must be in {{ $r.In }}",
					nil,
				)
				if !all { return err }
				errors = append(errors, err)
			}
		{{- else if $r.NotIn }}
			if _, ok := {{ lookup $f "NotInLookup" }}[a.GetTypeUrl()]; ok {
				err := {{GetAliasName}}ValidationError(
					{{.ErrCode}}, 
					"[{{ $f.GoName }}] type URL must not be in {{ $r.NotIn }}",
					nil,
				)
				if !all { return err }
				errors = append(errors, err)
			}
		{{- end }}
	}
`