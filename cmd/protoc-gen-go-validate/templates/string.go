package templates

const StrTpl = `
	{{ $f := .Field }}{{ $r := .Rules }}

	{{ if $r.GetIgnoreEmpty }}
		if {{ .GetAccessor }} != "" {
	{{ end }}

	{{ template "const" . }}
	{{ template "in" . }}

	{{ if or $r.Len (and $r.MinLen $r.MaxLen (eq $r.GetMinLen $r.GetMaxLen)) }}
		{{ if $r.Len }}
		if utf8.RuneCountInString({{ .GetAccessor }}) != {{ $r.GetLen }} {
			err := {{ err .Field "value length must be " $r.GetLen " runes" }}
			if !all { return err }
			errors = append(errors, err)
		{{ else }}
		if utf8.RuneCountInString({{ .GetAccessor }}) != {{ $r.GetMinLen }} {
			err := {{ err .Field "value length must be " $r.GetMinLen " runes" }}
			if !all { return err }
			errors = append(errors, err)
		{{ end }}
	}
	{{ else if $r.MinLen }}
		{{ if $r.MaxLen }}
			if l := utf8.RuneCountInString({{ .GetAccessor }}); l < {{ $r.GetMinLen }} || l > {{ $r.GetMaxLen }} {
				err := {{ err .Field "value length must be between " $r.GetMinLen " and " $r.GetMaxLen " runes, inclusive" }}
				if !all { return err }
				errors = append(errors, err)
			}
		{{ else }}
			if utf8.RuneCountInString({{ .GetAccessor }}) < {{ $r.GetMinLen }} {
				err := {{ err .Field "value length must be at least " $r.GetMinLen " runes" }}
				if !all { return err }
				errors = append(errors, err)
			}
		{{ end }}
	{{ else if $r.MaxLen }}
		if utf8.RuneCountInString({{ .GetAccessor }}) > {{ $r.GetMaxLen }} {
			err := {{ err .Field "value length must be at most " $r.GetMaxLen " runes" }}
			if !all { return err }
			errors = append(errors, err)
		}
	{{ end }}

	{{ if or $r.LenBytes (and $r.MinBytes $r.MaxBytes (eq $r.GetMinBytes $r.GetMaxBytes)) }}
		{{ if $r.LenBytes }}
			if len({{ .GetAccessor }}) != {{ $r.GetLenBytes }} {
				err := {{ err .Field "value length must be " $r.GetLenBytes " bytes" }}
				if !all { return err }
				errors = append(errors, err)
			}
		{{ else }}
			if len({{ .GetAccessor }}) != {{ $r.GetMinBytes }} {
				err := {{ err .Field "value length must be " $r.GetMinBytes " bytes" }}
				if !all { return err }
				errors = append(errors, err)
			}
		{{ end }}
	{{ else if $r.MinBytes }}
		{{ if $r.MaxBytes }}
			if l := len({{ .GetAccessor }}); l < {{ $r.GetMinBytes }} || l > {{ $r.GetMaxBytes }} {
					err := {{ err .Field "value length must be between " $r.GetMinBytes " and " $r.GetMaxBytes " bytes, inclusive" }}
					if !all { return err }
					errors = append(errors, err)
			}
		{{ else }}
			if len({{ .GetAccessor }}) < {{ $r.GetMinBytes }} {
				err := {{ err .Field "value length must be at least " $r.GetMinBytes " bytes" }}
				if !all { return err }
				errors = append(errors, err)
			}
		{{ end }}
	{{ else if $r.MaxBytes }}
		if len({{ .GetAccessor }}) > {{ $r.GetMaxBytes }} {
			err := {{ err .Field "value length must be at most " $r.GetMaxBytes " bytes" }}
			if !all { return err }
			errors = append(errors, err)
		}
	{{ end }}

	{{ if $r.Prefix }}
		if !strings.HasPrefix({{ .GetAccessor }}, {{ lit $r.GetPrefix }}) {
			err := {{ err .Field "value does not have prefix " (lit $r.GetPrefix) }}
			if !all { return err }
			errors = append(errors, err)
		}
	{{ end }}

	{{ if $r.Suffix }}
		if !strings.HasSuffix({{ .GetAccessor }}, {{ lit $r.GetSuffix }}) {
			err := {{ err .Field "value does not have suffix " (lit $r.GetSuffix) }}
			if !all { return err }
			errors = append(errors, err)
		}
	{{ end }}

	{{ if $r.Contains }}
		if !strings.Contains({{ .GetAccessor }}, {{ lit $r.GetContains }}) {
			err := {{ err .Field "value does not contain substring " (lit $r.GetContains) }}
			if !all { return err }
			errors = append(errors, err)
		}
	{{ end }}

	{{ if $r.NotContains }}
		if strings.Contains({{ .GetAccessor }}, {{ lit $r.GetNotContains }}) {
			err := {{ err .Field "value contains substring " (lit $r.GetNotContains) }}
			if !all { return err }
			errors = append(errors, err)
		}
	{{ end }}

	{{ if $r.GetIp }}
		if ip := net.ParseIP({{ .GetAccessor }}); ip == nil {
			err := {{ err .Field "value must be a valid IP address" }}
			if !all { return err }
			errors = append(errors, err)
		}
	{{ else if $r.GetIpv4 }}
		if ip := net.ParseIP({{ .GetAccessor }}); ip == nil || ip.To4() == nil {
			err := {{ err .Field "value must be a valid IPv4 address" }}
			if !all { return err }
			errors = append(errors, err)
		}
	{{ else if $r.GetIpv6 }}
		if ip := net.ParseIP({{ .GetAccessor }}); ip == nil || ip.To4() != nil {
			err := {{ err .Field "value must be a valid IPv6 address" }}
			if !all { return err }
			errors = append(errors, err)
		}
	{{ else if $r.GetEmail }}
		if err := _validateEmail({{ .GetAccessor }}); err != nil {
			err = {{ errCause .Field "err" "value must be a valid email address" }}
			if !all { return err }
			errors = append(errors, err)
		}
	{{ else if $r.GetHostname }}
		if err := _validateHostname({{ .GetAccessor }}); err != nil {
			err = {{ errCause .Field "err" "value must be a valid hostname" }}
			if !all { return err }
			errors = append(errors, err)
		}
	{{ else if $r.GetAddress }}
		if err := _validateHostname({{ .GetAccessor }}); err != nil {
			if ip := net.ParseIP({{ .GetAccessor }}); ip == nil {
				err := {{ err .Field "value must be a valid hostname, or ip address" }}
				if !all { return err }
				errors = append(errors, err)
			}
		}
	{{ else if $r.GetUri }}
		if uri, err := url.Parse({{ .GetAccessor }}); err != nil {
			err = {{ errCause .Field "err" "value must be a valid URI" }}
			if !all { return err }
			errors = append(errors, err)
		} else if !uri.IsAbs() {
			err := {{ err .Field "value must be absolute" }}
			if !all { return err }
			errors = append(errors, err)
		}
	{{ else if $r.GetUriRef }}
		if _, err := url.Parse({{ .GetAccessor }}); err != nil {
			err = {{ errCause .Field "err" "value must be a valid URI" }}
			if !all { return err }
			errors = append(errors, err)
		}
	{{ else if $r.GetUuid }}
		if err := _validateUuid({{ .GetAccessor }}); err != nil {
			err = {{ errCause .Field "err" "value must be a valid UUID" }}
			if !all { return err }
			errors = append(errors, err)
		}
	{{ end }}

	{{ if $r.Pattern }}
		if !{{ lookup .Field "Pattern" }}.MatchString({{ .GetAccessor }}) {
			err := {{ err .Field "value does not match regex pattern " (lit $r.GetPattern) }}
			if !all { return err }
			errors = append(errors, err)
		}
	{{ end }}

	{{ if $r.GetIgnoreEmpty }}
		}
	{{ end }}

`
