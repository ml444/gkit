package templates

// Embedded message validation.

const MessageTpl = `
	{{- $f := .Field }}{{ $r := .Rules }}
	{{- template "required" . }}

	{{- if .Rules.GetSkip }}
		// skipping validation for {{ .Field.GoName }}
	{{- else }}
		if all {
			switch v := interface{}({{ .GetAccessor }}).(type) {
				case interface{ ValidateAll() error }:
					if err := v.ValidateAll(); err != nil {
						errors = append(errors, {{GetAliasName}}ValidationError(
							{{.ErrCode}}, 
							"embedded message failed validation",
							err,
						))
					}
				case interface{ Validate() error }:
					{{- /* Support legacy validation for messages that were generated with a plugin version prior to existence of ValidateAll() */ -}}
					if err := v.Validate(); err != nil {
						errors = append(errors, {{GetAliasName}}ValidationError(
							{{.ErrCode}}, 
							"embedded message failed validation",
							err,
						))
					}
			}
		} else if v, ok := interface{}({{ .GetAccessor }}).(interface{ Validate() error }); ok {
			if err := v.Validate(); err != nil {
				return {{GetAliasName}}ValidationError(
					{{.ErrCode}}, 
					"embedded message failed validation",
					err,
				)
			}
		}
	{{- end }}
`
