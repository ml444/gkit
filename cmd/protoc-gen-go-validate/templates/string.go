package templates

const StrTpl = `
	{{- $f := .Field }}{{ $r := .Rules }}
	{{- if $r.GetIgnoreEmpty }}
		if {{ .GetAccessor }} != "" {
	{{- end -}}

	{{- template "const" . }}
	{{- template "in" . }}

	{{- if or $r.Len (and $r.MinLen $r.MaxLen (eq $r.GetMinLen $r.GetMaxLen)) }}
		{{- if $r.Len }}
		if utf8.RuneCountInString({{ .GetAccessor }}) != {{ $r.GetLen }} {
			err := {{GetAliasName}}ValidationError(
				{{.ErrCode}}, 
				"[{{ $f.GoName }}] value length must be {{ $r.GetLen }} runes", 
				nil,
			)
			if !all { return err }
			errors = append(errors, err)
		{{- else }}
		if utf8.RuneCountInString({{ .GetAccessor }}) != {{ $r.GetMinLen }} {
			err := {{GetAliasName}}ValidationError(
				{{.ErrCode}}, 
				"[{{ $f.GoName }}] value length must be {{ $r.GetMinLen }} runes", 
				nil,
			)
			if !all { return err }
			errors = append(errors, err)
		{{- end }}
	}
	{{- else if $r.MinLen }}
		{{- if $r.MaxLen }}
			if l := utf8.RuneCountInString({{ .GetAccessor }}); l < {{ $r.GetMinLen }} || l > {{ $r.GetMaxLen }} {
				err := {{GetAliasName}}ValidationError(
					{{.ErrCode}}, 
					"[{{ $f.GoName }}] value length must be between {{ $r.GetMinLen }} and {{ $r.GetMaxLen }} runes, inclusive",
					nil,
				)
				if !all { return err }
				errors = append(errors, err)
			}
		{{- else }}
			if utf8.RuneCountInString({{ .GetAccessor }}) < {{ $r.GetMinLen }} {
				err := {{GetAliasName}}ValidationError(
					{{.ErrCode}}, 
					"[{{ $f.GoName }}] value length must be at least {{ $r.GetMinLen }} runes",
					nil,
				)
				if !all { return err }
				errors = append(errors, err)
			}
		{{- end }}
	{{- else if $r.MaxLen }}
		if utf8.RuneCountInString({{ .GetAccessor }}) > {{ $r.GetMaxLen }} {
			err := {{GetAliasName}}ValidationError(
				{{.ErrCode}}, 
				"[{{ $f.GoName }}] value length must be at most {{ $r.GetMaxLen }} runes", 
				nil,
			)
			if !all { return err }
			errors = append(errors, err)
		}
	{{- end -}}

	{{- if or $r.LenBytes (and $r.MinBytes $r.MaxBytes (eq $r.GetMinBytes $r.GetMaxBytes)) }}
		{{- if $r.LenBytes }}
			if len({{ .GetAccessor }}) != {{ $r.GetLenBytes }} {
				err := {{GetAliasName}}ValidationError(
					{{.ErrCode}}, 
					"[{{ $f.GoName }}] value length must be {{ $r.GetLenBytes }} bytes", 
					nil,
				)
				if !all { return err }
				errors = append(errors, err)
			}
		{{- else }}
			if len({{ .GetAccessor }}) != {{ $r.GetMinBytes }} {
				err := {{GetAliasName}}ValidationError(
					{{.ErrCode}}, 
					"[{{ $f.GoName }}] value length must be {{ $r.GetMinBytes }} bytes", 
					nil,
				)
				if !all { return err }
				errors = append(errors, err)
			}
		{{- end }}
	{{- else if $r.MinBytes }}
		{{- if $r.MaxBytes }}
			if l := len({{ .GetAccessor }}); l < {{ $r.GetMinBytes }} || l > {{ $r.GetMaxBytes }} {
				err := {{GetAliasName}}ValidationError(
					{{.ErrCode}}, 
					"[{{ $f.GoName }}] value length must be between {{ $r.GetMinBytes }} and {{ $r.GetMaxBytes }} bytes, inclusive", 
					nil,
				)
				if !all { return err }
				errors = append(errors, err)
			}
		{{- else }}
			if len({{ .GetAccessor }}) < {{ $r.GetMinBytes }} {
				err := {{GetAliasName}}ValidationError(	
					{{.ErrCode}}, 
					"[{{ $f.GoName }}] value length must be at least {{ $r.GetMinBytes }} bytes", 
					nil,
				)
				if !all { return err }
				errors = append(errors, err)
			}
		{{- end }}
	{{- else if $r.MaxBytes }}
		if len({{ .GetAccessor }}) > {{ $r.GetMaxBytes }} {
			err := {{GetAliasName}}ValidationError(
				{{.ErrCode}}, 
				"[{{ $f.GoName }}] value length must be at most {{ $r.GetMaxBytes }} bytes", 
				nil,
			)
			if !all { return err }
			errors = append(errors, err)
		}
	{{- end -}}

	{{- if $r.Prefix }}
		if !strings.HasPrefix({{ .GetAccessor }}, {{ lit $r.GetPrefix }}) {
			err := {{GetAliasName}}ValidationError(
				{{.ErrCode}}, 
				fmt.Sprintf("[{{ $f.GoName }}] value does not have prefix '%s'", {{ lit $r.GetPrefix }}), 
				nil,
			)
			if !all { return err }
			errors = append(errors, err)
		}
	{{- end -}}

	{{- if $r.Suffix }}
		if !strings.HasSuffix({{ .GetAccessor }}, {{ lit $r.GetSuffix }}) {
			err := {{GetAliasName}}ValidationError(
				{{.ErrCode}}, 
				fmt.Sprintf("[{{ $f.GoName }}] value does not have suffix '%s'", {{ lit $r.GetSuffix }}), 
				nil,
			)
			if !all { return err }
			errors = append(errors, err)
		}
	{{- end -}}

	{{- if $r.Contains }}
		if !strings.Contains({{ .GetAccessor }}, {{ lit $r.GetContains }}) {
			err := {{GetAliasName}}ValidationError(
				{{.ErrCode}}, 
				fmt.Sprintf("[{{ $f.GoName }}] value does not contain substring '%s'", {{ lit $r.GetContains }}), 
				nil,
			)
			if !all { return err }
			errors = append(errors, err)
		}
	{{- end -}}

	{{- if $r.NotContains }}
		if strings.Contains({{ .GetAccessor }}, {{ lit $r.GetNotContains }}) {
			err := {{GetAliasName}}ValidationError(
				{{.ErrCode}}, 
				fmt.Sprintf("[{{ $f.GoName }}] value contains substring '%s'", {{ lit $r.GetNotContains }}), 
				nil,
			)
			if !all { return err }
			errors = append(errors, err)
		}
	{{- end -}}

	{{- if $r.GetIp }}
		if ip := net.ParseIP({{ .GetAccessor }}); ip == nil {
			err := {{GetAliasName}}ValidationError(
				{{.ErrCode}}, 
				"[{{ $f.GoName }}] value must be a valid IP address",
				nil,
			)
			if !all { return err }
			errors = append(errors, err)
		}
	{{- else if $r.GetIpv4 }}
		if ip := net.ParseIP({{ .GetAccessor }}); ip == nil || ip.To4() == nil {
			err := {{GetAliasName}}ValidationError(
				{{.ErrCode}}, 
				"[{{ $f.GoName }}] value must be a valid IPv4 address",
				nil,
			)
			if !all { return err }
			errors = append(errors, err)
		}
	{{- else if $r.GetIpv6 }}
		if ip := net.ParseIP({{ .GetAccessor }}); ip == nil || ip.To4() != nil {
			err := {{GetAliasName}}ValidationError(
				{{.ErrCode}}, 
				"[{{ $f.GoName }}] value must be a valid IPv6 address",
				nil,
			)
			if !all { return err }
			errors = append(errors, err)
		}
	{{- else if $r.GetEmail }}
		if e := {{GetAliasName}}_validateEmail({{ .GetAccessor }}); e != nil {
			err := {{GetAliasName}}ValidationError(
				{{.ErrCode}}, 
				"[{{ $f.GoName }}] value must be a valid email address",
				e,
			)
			if !all { return err }
			errors = append(errors, err)
		}
	{{- else if $r.GetHostname }}
		if e := {{GetAliasName}}_validateHostname({{ .GetAccessor }}); e != nil {
			err := {{GetAliasName}}ValidationError(
				{{.ErrCode}}, 
				"[{{ $f.GoName }}] value must be a valid hostname",
				e,
			)
			if !all { return err }
			errors = append(errors, err)
		}
	{{- else if $r.GetAddress }}
		if e := {{GetAliasName}}_validateHostname({{ .GetAccessor }}); e != nil {
			if ip := net.ParseIP({{ .GetAccessor }}); ip == nil {
				err := {{GetAliasName}}ValidationError(
					{{.ErrCode}}, 
					"[{{ $f.GoName }}] value must be a valid hostname, or ip address",
					e,
				)
				if !all { return err }
				errors = append(errors, err)
			}
		}
	{{- else if $r.GetUri }}
		if uri, e := url.Parse({{ .GetAccessor }}); e != nil {
			err := {{GetAliasName}}ValidationError(
				{{.ErrCode}}, 
				"[{{ $f.GoName }}] value must be a valid URI",
				e,
			)
			if !all { return err }
			errors = append(errors, err)
		} else if !uri.IsAbs() {
			err := {{GetAliasName}}ValidationError(
				{{.ErrCode}}, 
				"[{{ $f.GoName }}] value must be absolute",
				nil,
			)
			if !all { return err }
			errors = append(errors, err)
		}
	{{- else if $r.GetUriRef }}
		if _, e := url.Parse({{ .GetAccessor }}); e != nil {
			err := {{GetAliasName}}ValidationError(
				{{.ErrCode}}, 
				"[{{ $f.GoName }}] value must be a valid URI",
				e,
			)
			if !all { return err }
			errors = append(errors, err)
		}
	{{- else if $r.GetUuid }}
		if e := {{GetAliasName}}_validateUuid({{ .GetAccessor }}); e != nil {
			err := {{GetAliasName}}ValidationError(
				{{.ErrCode}}, 
				"[{{ $f.GoName }}] value must be a valid UUID",
				e,
			)
			if !all { return err }
			errors = append(errors, err)
		}
	{{- else if $r.GetWellKnownRegex }}
		if !{{ lookup $f "WellKnownPattern" }}.MatchString({{ .GetAccessor }}) {
			err := {{GetAliasName}}ValidationError(
				{{.ErrCode}}, 
				"[{{ $f.GoName }}] {{ wellKnownRegexErrMsg $r }}",
				nil,
			)
			if !all { return err }
			errors = append(errors, err)
		}
	{{- end -}}

	{{- if $r.Pattern }}
		if !{{ lookup .Field "Pattern" }}.MatchString({{ .GetAccessor }}) {
			err := {{GetAliasName}}ValidationError(
				{{.ErrCode}}, 
				fmt.Sprintf("[{{ $f.GoName }}] value does not match regex pattern '%s'", {{ lit $r.GetPattern }}), 
				nil,
			)
			if !all { return err }
			errors = append(errors, err)
		}
	{{- end -}}

	{{- if $r.GetIgnoreEmpty }}
		}
	{{- end }}
`
