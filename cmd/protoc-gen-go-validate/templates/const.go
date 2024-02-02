package templates

const ConstTpl = `{{ $f := .Field }}{{ $r := .Rules }}
	{{ if $r.Const }}
		if {{ .GetAccessor }} != {{ lit $r.GetConst }} {
			{{- if isEnum $f }}
			err := {{ err .Field "value must equal " (enumVal $f $r.GetConst) }}
			{{- else }}
			err := {{ err .Field "value must equal " $r.GetConst }}
			{{- end }}
			if !all { return err }
			errors = append(errors, err)
		}
	{{ end }}
`
