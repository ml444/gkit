package templates

const RepTpl = `
	{{ $f := .Field }}{{ $r := .Rules }}

	{{ if $r.GetIgnoreEmpty }}
		if len({{ .GetAccessor }}) > 0 {
	{{ end }}

	{{ if $r.GetMinItems }}
		{{ if eq $r.GetMinItems $r.GetMaxItems }}
			if len({{ .GetAccessor }}) != {{ $r.GetMinItems }} {
				err := {{ err .Field "value must contain exactly " $r.GetMinItems " item(s)" }}
				if !all { return err }
				errors = append(errors, err)
			}
		{{ else if $r.MaxItems }}
			if l := len({{ .GetAccessor }}); l < {{ $r.GetMinItems }} || l > {{ $r.GetMaxItems }} {
				err := {{ err .Field "value must contain between " $r.GetMinItems " and " $r.GetMaxItems " items, inclusive" }}
				if !all { return err }
				errors = append(errors, err)
			}
		{{ else }}
			if len({{ .GetAccessor }}) < {{ $r.GetMinItems }} {
				err := {{ err .Field "value must contain at least " $r.GetMinItems " item(s)" }}
				if !all { return err }
				errors = append(errors, err)
			}
		{{ end }}
	{{ else if $r.MaxItems }}
		if len({{ .GetAccessor }}) > {{ $r.GetMaxItems }} {
			err := {{ err .Field "value must contain no more than " $r.GetMaxItems " item(s)" }}
			if !all { return err }
			errors = append(errors, err)
		}
	{{ end }}

	{{ if $r.GetUnique }}
		{{ lookup $f "Unique" }} := {{ if isBytes $f.Desc -}}
			make(map[string]struct{}, len({{ .GetAccessor }}))
		{{ else -}}
			make(map[{{ .Type }}]struct{}, len({{ .GetAccessor }}))
		{{ end -}}
	{{ end }}

	{{ if or $r.GetUnique $r.GetItems (has .Rules "In") }}
		for idx, item := range {{ .GetAccessor }} {
			_, _ = idx, item
			{{ if $r.GetUnique }}
				if _, exists := {{ lookup $f "Unique" }}[{{ if isBytes $f.Desc }}string(item){{ else }}item{{ end }}]; exists {
					err := {{ errIdx .Field "idx" "repeated value must contain unique items" }}
					if !all { return err }
					errors = append(errors, err)
				} else {
					{{ lookup $f "Unique" }}[{{ if isBytes $f.Desc }}string(item){{ else }}item{{ end }}] = struct{}{}
				}
			{{ end }}

			{{- $elem := (.Elem "item") }}
			{{ render $elem.TmplName $elem }}
		}
	{{ end }}

	{{ if $r.GetIgnoreEmpty }}
		}
	{{ end }}
`
