package templates

const BytesTpl = `
	{{ $f := .Field }}{{ $r := .Rules }}

	{{ if $r.GetIgnoreEmpty }}
		if len({{ .GetAccessor }}) > 0 {
	{{ end }}

	{{ if or $r.Len (and $r.MinLen $r.MaxLen (eq $r.GetMinLen $r.GetMaxLen)) }}
		{{ if $r.Len }}
			if len({{ .GetAccessor }}) != {{ $r.GetLen }} {
				err := {{ err .Field "value length must be " $r.GetLen " bytes" }}
				if !all { return err }
				errors = append(errors, err)
			}
		{{ else }}
			if len({{ .GetAccessor }}) != {{ $r.GetMinLen }} {
				err := {{ err .Field "value length must be " $r.GetMinLen " bytes" }}
				if !all { return err }
				errors = append(errors, err)
			}
		{{ end }}
	{{ else if $r.MinLen }}
		{{ if $r.MaxLen }}
			if l := len({{ .GetAccessor }}); l < {{ $r.GetMinLen }} || l > {{ $r.GetMaxLen }} {
				err := {{ err .Field "value length must be between " $r.GetMinLen " and " $r.GetMaxLen " bytes, inclusive" }}
				if !all { return err }
				errors = append(errors, err)
			}
		{{ else }}
			if len({{ .GetAccessor }}) < {{ $r.GetMinLen }} {
				err := {{ err .Field "value length must be at least " $r.GetMinLen " bytes" }}
				if !all { return err }
				errors = append(errors, err)
			}
		{{ end }}
	{{ else if $r.MaxLen }}
		if len({{ .GetAccessor }}) > {{ $r.GetMaxLen }} {
			err := {{ err .Field "value length must be at most " $r.GetMaxLen " bytes" }}
			if !all { return err }
			errors = append(errors, err)
		}
	{{ end }}

	{{ if $r.Prefix }}
		if !bytes.HasPrefix({{ .GetAccessor }}, {{ lit $r.GetPrefix }}) {
			err := {{ err .Field "value does not have prefix " (byteStr $r.GetPrefix) }}
			if !all { return err }
			errors = append(errors, err)
		}
	{{ end }}

	{{ if $r.Suffix }}
		if !bytes.HasSuffix({{ .GetAccessor }}, {{ lit $r.GetSuffix }}) {
			err := {{ err .Field "value does not have suffix " (byteStr $r.GetSuffix) }}
			if !all { return err }
			errors = append(errors, err)
		}
	{{ end }}

	{{ if $r.Contains }}
		if !bytes.Contains({{ .GetAccessor }}, {{ lit $r.GetContains }}) {
			err := {{ err .Field "value does not contain " (byteStr $r.GetContains) }}
			if !all { return err }
			errors = append(errors, err)
		}
	{{ end }}

	{{ if $r.In }}
		if _, ok := {{ lookup $f "InLookup" }}[string({{ .GetAccessor }})]; !ok {
			err := {{ err .Field "value must be in list " $r.In }}
			if !all { return err }
			errors = append(errors, err)
		}
	{{ else if $r.NotIn }}
		if _, ok := {{ lookup $f "NotInLookup" }}[string({{ .GetAccessor }})]; ok {
			err := {{ err .Field "value must not be in list " $r.NotIn }}
			if !all { return err }
			errors = append(errors, err)
		}
	{{ end }}

	{{ if $r.Const }}
		if !bytes.Equal({{ .GetAccessor }}, {{ lit $r.Const }}) {
			err := {{ err .Field "value must equal " $r.Const }}
			if !all { return err }
			errors = append(errors, err)
		}
	{{ end }}

	{{ if $r.GetIp }}
		if ip := net.IP({{ .GetAccessor }}); ip.To16() == nil {
			err := {{ err .Field "value must be a valid IP address" }}
			if !all { return err }
			errors = append(errors, err)
		}
	{{ else if $r.GetIpv4 }}
		if ip := net.IP({{ .GetAccessor }}); ip.To4() == nil {
			err := {{ err .Field "value must be a valid IPv4 address" }}
			if !all { return err }
			errors = append(errors, err)
		}
	{{ else if $r.GetIpv6 }}
		if ip := net.IP({{ .GetAccessor }}); ip.To16() == nil || ip.To4() != nil {
			err := {{ err .Field "value must be a valid IPv6 address" }}
			if !all { return err }
			errors = append(errors, err)
		}
	{{ end }}

	{{ if $r.Pattern }}
	if !{{ lookup $f "Pattern" }}.Match({{ .GetAccessor }}) {
		err := {{ err .Field "value does not match regex pattern " (lit $r.GetPattern) }}
		if !all { return err }
		errors = append(errors, err)
	}
	{{ end }}

	{{ if $r.GetIgnoreEmpty }}
		}
	{{ end }}
`
