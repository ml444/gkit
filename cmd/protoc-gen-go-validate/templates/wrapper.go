package templates

const WrapperTpl = `
	{{ $f := .Field }}{{ $r := .Rules }}

	if wrapper := {{ .GetAccessor }}; wrapper != nil {
		{{ $ctx := .Unwrap "wrapper" -}}
		{{ render $ctx.TmplName $ctx }}
	} {{ if .Required }} else {
		err := {{GetAliasName}}ValidationError(
			{{.ErrCode}}, 
			"[{{ $f.GoName }}] value is required and must not be nil.",
			nil,
		)
		if !all { return err }
		errors = append(errors, err)
	} {{ end }}
`
