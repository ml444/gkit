package templates

const EnumTpl = `
		{{ $f := .Field }}{{ $r := .Rules }}
		{{ template "const" . }}
		{{ template "in" . }}

		{{ if $r.GetDefinedOnly }}
			if _, ok := {{ enumName .Desc.Enum }}_name[int32({{ .GetAccessor }})]; !ok {
				err := {{ err .Field "value must be one of the defined enum values" }}
				if !all { return err }
				errors = append(errors, err)
			}
		{{ end }}
`
