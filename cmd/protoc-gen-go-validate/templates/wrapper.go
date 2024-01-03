package templates

const WrapperTpl = `
	{{ $f := .Field }}{{ $r := .Rules }}

	if wrapper := {{ accessor . }}; wrapper != nil {
		{{ $ctx := .Unwrap "wrapper" -}}
		{{ render $ctx.TmplName $ctx }}
	} {{ if .Required }} else {
		err := {{ err .Field "value is required and must not be nil." }}
		if !all { return err }
		errors = append(errors, err)
	} {{ end }}
`
