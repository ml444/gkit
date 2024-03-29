package templates

const InTpl = `{{ $f := .Field }}{{ $r := .Rules }}
	{{ if $r.In }}
		if _, ok := {{ lookup $f "InLookup" }}[{{ .GetAccessor }}]; !ok {
			{{- if isEnum $f }}
			err := {{ err .Field "value must be in list " (enumList $f $r.In) }}
			{{- else }}
			err := {{ err .Field "value must be in list " $r.In }}
			{{- end }}
			if !all { return err }
			errors = append(errors, err)
		}
	{{ else if $r.NotIn }}
		if _, ok := {{ lookup $f "NotInLookup" }}[{{ .GetAccessor }}]; ok {
			{{- if isEnum $f }}
			err := {{ err .Field "value must not be in list " (enumList $f $r.NotIn) }}
			{{- else }}
			err := {{ err .Field "value must not be in list " $r.NotIn }}
			{{- end }}
			if !all { return err }
			errors = append(errors, err)
		}
	{{ end }}
`
