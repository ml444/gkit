package templates

const EnumTpl = `
		{{ $f := .Field }}{{ $r := .Rules }}
		{{ template "const" . }}
		{{ template "in" . }}

		{{ if $r.GetDefinedOnly }}
			if _, ok := {{ enumName .Desc.Enum true }}_name[int32({{ accessor . }})]; !ok {
				err := {{ err .Field "value must be one of the defined enum values" }}
				if !all { return err }
				errors = append(errors, err)
			}
		{{ end }}
`
