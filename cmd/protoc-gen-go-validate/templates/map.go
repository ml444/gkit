package templates

const MapTpl = `
	{{ $data := . }}{{ $f := .Field }}{{ $r := .Rules }}

	{{- if $r.GetIgnoreEmpty }}
		if len({{ .GetAccessor }}) > 0 {
	{{- end }}

	{{- if $r.GetMinPairs }}
		{{- if eq $r.GetMinPairs $r.GetMaxPairs }}
			if len({{ .GetAccessor }}) != {{ $r.GetMinPairs }} {
				err := {{GetAliasName}}ValidationError(
					{{.ErrCode}}, 
					"[{{ $f.GoName }}] value must contain exactly {{ $r.GetMinPairs }} pair(s)", 
					nil,
				)
				if !all { return err }
				errors = append(errors, err)
			}
		{{- else if $r.MaxPairs }}
			if l := len({{ .GetAccessor }}); l < {{ $r.GetMinPairs }} || l > {{ $r.GetMaxPairs }} {
				err := {{GetAliasName}}ValidationError(
					{{.ErrCode}}, 
					"[{{ $f.GoName }}] value must contain between {{ $r.GetMinPairs }} and {{ $r.GetMaxPairs }} pairs, inclusive",
					nil,
				)
				if !all { return err }
				errors = append(errors, err)
			}
		{{- else }}
			if len({{ .GetAccessor }}) < {{ $r.GetMinPairs }} {
				err := {{GetAliasName}}ValidationError(
					{{.ErrCode}}, 
					"[{{ $f.GoName }}] value must contain at least {{ $r.GetMinPairs }} pair(s)",
					nil,
				)
				if !all { return err }
				errors = append(errors, err)
			}
		{{- end }}
	{{- else if $r.MaxPairs }}
		if len({{ .GetAccessor }}) > {{ $r.GetMaxPairs }} {
			err := {{GetAliasName}}ValidationError(
				{{.ErrCode}}, 
				"[{{ $f.GoName }}] value must contain no more than {{ $r.GetMaxPairs }} pair(s)",
				nil,
			)
			if !all { return err }
			errors = append(errors, err)
		}
	{{- end }}

	{{- if or $r.GetNoSparse $r.GetValues $r.GetKeys }}
		{
			if all {
				sorted_keys := make([]{{ inKind .Desc.MapKey.Kind }}, len({{ .GetAccessor }}))
				i := 0
				for key := range {{ .GetAccessor }} {
					sorted_keys[i] = key
					i++
				}
				sort.Slice(sorted_keys, func (i, j int) bool { return sorted_keys[i] < sorted_keys[j] })
				for _, key := range sorted_keys {
					val := {{ .GetAccessor }}[key]
					_ = val
					{{ template "mapPairBody" $data }}
				}
			} else {
				for key, val := range {{ .GetAccessor }} {
					_, _ = key, val
					{{ template "mapPairBody" $data }}
				}
			}
		}
	{{ end }}

	{{ if $r.GetIgnoreEmpty }}
		}
	{{ end }}

`

const MapPairBodyTpl = `
	{{ $f := .Field }}{{ $r := .Rules }}
	{{- if $r.GetNoSparse }}
		if val == nil {
			err := {{GetAliasName}}ValidationError(
				{{.ErrCode}}, 
				fmt.Sprintf("[{{ $f.GoName }}] key:%v value cannot be sparse, all pairs must be non-nil", key),
				nil,
			)
			if !all { return err }
			errors = append(errors, err)
		}
	{{- end }}
	{{- $ctx :=(.MapTypeName $f $f.Desc.MapKey $r.GetKeys) }} 
	{{- $ctx = $ctx.SetAccessor "key" }}
	{{- render $ctx.TmplName $ctx }}

	{{- $ctx :=(.MapTypeName $f $f.Desc.MapValue $r.GetValues) }} 
	{{- $ctx = $ctx.SetAccessor "val" }}
	{{- render $ctx.TmplName $ctx }}
`
