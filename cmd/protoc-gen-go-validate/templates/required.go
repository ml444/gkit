package templates

const RequiredTpl = `
	{{- if .Rules.GetRequired }}
		if {{ .GetAccessor }} == nil {
			err := {{GetAliasName}}ValidationError(
				{{.ErrCode}}, 
				"[{{ .Field.GoName }}] value is required",
				nil,
			)
			if !all { return err }
			errors = append(errors, err)
		}
	{{- end }}
`
