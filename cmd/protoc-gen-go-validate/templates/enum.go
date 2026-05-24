package templates

const EnumTpl = `
		{{- $f := .Field }}{{ $r := .Rules }}
		{{- template "const" . }}
		{{- template "in" . }}

		{{- if $r.GetDefinedOnly }}
			if _, ok := {{ enumName .Desc.Enum }}_name[int32({{ .GetAccessor }})]; !ok {
				err := {{GetAliasName}}ValidationError(
					{{.ErrCode}}, 
					"[{{ $f.GoName }}] value must be one of the defined enum values",
					nil,
				)
				if !all { return err }
				errors = append(errors, err)
			}
		{{- end }}
`
