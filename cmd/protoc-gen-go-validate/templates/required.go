package templates

const RequiredTpl = `
	{{ if .Rules.GetRequired }}
		if {{ .GetAccessor }} == nil {
			err := {{ err .Field "value is required" }}
			if !all { return err }
			errors = append(errors, err)
		}
	{{ end }}
`
