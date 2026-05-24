package templates

const BytesTpl = `
	{{- $f := .Field }}{{ $r := .Rules }}

	{{- if $r.GetIgnoreEmpty }}
		if len({{ .GetAccessor }}) > 0 {
	{{- end -}}

	{{- if or $r.Len (and $r.MinLen $r.MaxLen (eq $r.GetMinLen $r.GetMaxLen)) }}
		{{- if $r.Len }}
			if len({{ .GetAccessor }}) != {{ $r.GetLen }} {
				err := {{GetAliasName}}ValidationError(
					{{.ErrCode}}, 
					"[{{ $f.GoName }}] value length must be {{ $r.GetLen }} bytes", 
					nil,
				)
				if !all { return err }
				errors = append(errors, err)
			}
		{{- else }}
			if len({{ .GetAccessor }}) != {{ $r.GetMinLen }} {
				err := {{GetAliasName}}ValidationError(
					{{.ErrCode}}, 
					"[{{ $f.GoName }}] value length must be {{ $r.GetMinLen }} bytes", 
					nil,
				)
				if !all { return err }
				errors = append(errors, err)
			}
		{{- end }}
	{{- else if $r.MinLen }}
		{{- if $r.MaxLen }}
			if l := len({{ .GetAccessor }}); l < {{ $r.GetMinLen }} || l > {{ $r.GetMaxLen }} {
				err := {{GetAliasName}}ValidationError(
					{{.ErrCode}}, 
					"[{{ $f.GoName }}] value length must be between {{ $r.GetMinLen }} and {{ $r.GetMaxLen }} bytes, inclusive",
					nil,
				)
				if !all { return err }
				errors = append(errors, err)
			}
		{{- else }}
			if len({{ .GetAccessor }}) < {{ $r.GetMinLen }} {
				err := {{GetAliasName}}ValidationError(
					{{.ErrCode}}, 
					"[{{ $f.GoName }}] value length must be at least {{ $r.GetMinLen }} bytes", 
					nil,
				)
				if !all { return err }
				errors = append(errors, err)
			}
		{{- end }}
	{{- else if $r.MaxLen }}
		if len({{ .GetAccessor }}) > {{ $r.GetMaxLen }} {
			err := {{GetAliasName}}ValidationError(
				{{.ErrCode}}, 
				"[{{ $f.GoName }}] value length must be at most {{ $r.GetMaxLen }} bytes", 
				nil,
			)
			if !all { return err }
			errors = append(errors, err)
		}
	{{- end -}}

	{{- if $r.Prefix }}
		if !bytes.HasPrefix({{ .GetAccessor }}, {{ lit $r.GetPrefix }}) {
			err := {{GetAliasName}}ValidationError(
				{{.ErrCode}}, 
				fmt.Sprintf("[{{ $f.GoName }}] value does not have prefix '%s'", {{ byteStr $r.GetPrefix }}),
				nil,
			)
			if !all { return err }
			errors = append(errors, err)
		}
	{{- end -}}

	{{- if $r.Suffix }}
		if !bytes.HasSuffix({{ .GetAccessor }}, {{ lit $r.GetSuffix }}) {
			err := {{GetAliasName}}ValidationError(
				{{.ErrCode}}, 
				fmt.Sprintf("[{{ $f.GoName }}] value does not have suffix '%s'", {{ byteStr $r.GetSuffix }}),
				nil,
			)
			if !all { return err }
			errors = append(errors, err)
		}
	{{- end -}}

	{{- if $r.Contains }}
		if !bytes.Contains({{ .GetAccessor }}, {{ lit $r.GetContains }}) {
			err := {{GetAliasName}}ValidationError(
				{{.ErrCode}}, 
				fmt.Sprintf("[{{ $f.GoName }}] value does not contain '%s'", {{ byteStr $r.GetContains }}),
				nil,
			)
			if !all { return err }
			errors = append(errors, err)
		}
	{{- end -}}

	{{ template "in" . }}

	{{- if $r.Const }}
		if !bytes.Equal({{ .GetAccessor }}, {{ lit $r.Const }}) {
			err := {{GetAliasName}}ValidationError(
				{{.ErrCode}}, 
				"[{{ $f.GoName }}] value must equal {{ $r.Const }}",
				nil,
			)
			if !all { return err }
			errors = append(errors, err)
		}
	{{- end -}}

	{{- if $r.GetIp }}
		if ip := net.IP({{ .GetAccessor }}); ip.To16() == nil {
			err := {{GetAliasName}}ValidationError(
				{{.ErrCode}}, 
				"[{{ $f.GoName }}] value must be a valid IP address",
				nil,
			)
			if !all { return err }
			errors = append(errors, err)
		}
	{{- else if $r.GetIpv4 }}
		if ip := net.IP({{ .GetAccessor }}); ip.To4() == nil {
			err := {{GetAliasName}}ValidationError(
				{{.ErrCode}}, 
				"[{{ $f.GoName }}] value must be a valid IPv4 address",
				nil,
			)
			if !all { return err }
			errors = append(errors, err)
		}
	{{- else if $r.GetIpv6 }}
		if ip := net.IP({{ .GetAccessor }}); ip.To16() == nil || ip.To4() != nil {
			err := {{GetAliasName}}ValidationError(
				{{.ErrCode}}, 
				"[{{ $f.GoName }}] value must be a valid IPv6 address",
				nil,
			)
			if !all { return err }
			errors = append(errors, err)
		}
	{{- end -}}

	{{- if $r.Pattern }}
	if !{{ lookup $f "Pattern" }}.Match({{ .GetAccessor }}) {
		err := {{GetAliasName}}ValidationError(
			{{.ErrCode}}, 
			fmt.Sprintf("[{{ $f.GoName }}] value does not match regex pattern '%s'", {{ lit $r.GetPattern }}),
			nil,
		)
		if !all { return err }
		errors = append(errors, err)
	}
	{{- end }}
	{{- if $r.GetIgnoreEmpty }}
		}
	{{- end }}
`
