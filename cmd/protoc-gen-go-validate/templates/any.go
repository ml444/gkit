package templates

const AnyTpl = `{{ $f := .Field }}{{ $r := .Rules }}
	{{ template "required" . }}

	if a := {{ .GetAccessor }}; a != nil {
		{{ if $r.In }}
			if _, ok := {{ lookup $f "InLookup" }}[a.GetTypeUrl()]; !ok {
				err := {{ err .Field "type URL must be in list " $r.In }}
				if !all { return err }
				errors = append(errors, err)
			}
		{{ else if $r.NotIn }}
			if _, ok := {{ lookup $f "NotInLookup" }}[a.GetTypeUrl()]; ok {
				err := {{ err .Field "type URL must not be in list " $r.NotIn }}
				if !all { return err }
				errors = append(errors, err)
			}
		{{ end }}
	}
`
