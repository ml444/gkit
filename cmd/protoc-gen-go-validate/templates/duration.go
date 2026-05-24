package templates

const DurationTpl = `{{ $f := .Field }}{{ $r := .Rules }}
	{{ template "required" . }}

	{{ if or $r.In $r.NotIn $r.Lt $r.Lte $r.Gt $r.Gte $r.Const }}
		if d := {{ .GetAccessor }}; d != nil {
			dur, err := d.AsDuration(), d.CheckValid()
			if err != nil {
				err := {{GetAliasName}}ValidationError(
					{{.ErrCode}},
					"[{{ $f.GoName }}] value is not a valid duration",
					nil,
				)
				if !all { return err }
				errors = append(errors, err)
			} else {
				{{ template "durationcmp" . }}
			}
		}
	{{ end }}
`
const DurationcmpTpl = `{{- $f := .Field }}{{ $r := .Rules }}
			{{- if $r.Const }}
				if dur != {{ durLit $r.Const }} {
					err := {{GetAliasName}}ValidationError(
						{{.ErrCode}}, 
						"[{{ $f.GoName }}] value must equal {{ durStr $r.Const }}",
						nil,
					)
					if !all { return err }
					errors = append(errors, err)
				}
			{{- end }}

			{{- if $r.Lt }}  lt  := {{ durLit $r.Lt }};  {{ end }}
			{{- if $r.Lte }} lte := {{ durLit $r.Lte }}; {{ end }}
			{{- if $r.Gt }}  gt  := {{ durLit $r.Gt }};  {{ end }}
			{{- if $r.Gte }} gte := {{ durLit $r.Gte }}; {{ end }}

			{{- if $r.Lt }}
				{{- if $r.Gt }}
					{{- if durGt $r.GetLt $r.GetGt }}
						if dur <= gt || dur >= lt {
							err := {{GetAliasName}}ValidationError(
								{{.ErrCode}}, 
								"[{{ $f.GoName }}] value must be inside range ({{durStr $r.GetGt}}, {{durStr $r.GetLt}})",
								nil,
							)
							if !all { return err }
							errors = append(errors, err)
						}
					{{- else }}
						if dur >= lt && dur <= gt {
							err := {{GetAliasName}}ValidationError(
								{{.ErrCode}}, 
								"[{{ $f.GoName }}] value must be outside range [{{durStr $r.GetLt}}, {{durStr $r.GetGt}}]",
								nil,
							)
							if !all { return err }
							errors = append(errors, err)
						}
					{{- end }}
				{{- else if $r.Gte }}
					{{- if durGt $r.GetLt $r.GetGte }}
						if dur < gte || dur >= lt {
							err := {{GetAliasName}}ValidationError(
								{{.ErrCode}}, 
								"[{{ $f.GoName }}] value must be inside range [{{durStr $r.GetGte}}, {{durStr $r.GetLt}})",
								nil,
							)
							if !all { return err }
							errors = append(errors, err)
						}
					{{- else }}
						if dur >= lt && dur < gte {
							err := {{GetAliasName}}ValidationError(
								{{.ErrCode}}, 
								"[{{ $f.GoName }}] value must be outside range [{{durStr $r.GetLt}}, {{durStr $r.GetGte}})",
								nil,
							)
							if !all { return err }
							errors = append(errors, err)
						}
					{{- end }}
				{{- else }}
					if dur >= lt {
						err := {{GetAliasName}}ValidationError(
							{{.ErrCode}}, 
							"[{{ $f.GoName }}] value must be less than {{durStr $r.GetLt}}",
							nil,
						)
						if !all { return err }
						errors = append(errors, err)
					}
				{{- end }}
			{{- else if $r.Lte }}
				{{- if $r.Gt }}
					{{- if durGt $r.GetLte $r.GetGt }}
						if dur <= gt || dur > lte {
							err := {{GetAliasName}}ValidationError(
								{{.ErrCode}}, 
								"[{{ $f.GoName }}] value must be inside range ({{durStr $r.GetGt}}, {{durStr $r.GetLte}}]",
								nil,
							)
							if !all { return err }
							errors = append(errors, err)
						}
					{{- else }}
						if dur > lte && dur <= gt {
							err := {{GetAliasName}}ValidationError(
								{{.ErrCode}}, 
								"[{{ $f.GoName }}] value must be outside range ({{durStr $r.GetLte}}, {{durStr $r.GetGt}}]",
								nil,
							)
							if !all { return err }
							errors = append(errors, err)
						}
					{{- end }}
				{{- else if $r.Gte }}
					{{- if durGt $r.GetLte $r.GetGte }}
						if dur < gte || dur > lte {
							err := {{GetAliasName}}ValidationError(
								{{.ErrCode}}, 
								"[{{ $f.GoName }}] value must be inside range [{{durStr $r.GetGte}}, {{durStr $r.GetLte}}]",
								nil,
							)
							if !all { return err }
							errors = append(errors, err)
						}
					{{- else }}
						if dur > lte && dur < gte {
							err := {{GetAliasName}}ValidationError(
								{{.ErrCode}}, 
								"[{{ $f.GoName }}] value must be outside range ({{durStr $r.GetLte}}, {{durStr $r.GetGte}})",
								nil,
							)
							if !all { return err }
							errors = append(errors, err)
						}
					{{- end }}
				{{- else }}
					if dur > lte {
						err := {{GetAliasName}}ValidationError(
							{{.ErrCode}}, 
							"[{{ $f.GoName }}] value must be less than or equal to {{durStr $r.GetLte}}",
							nil,
						)
						if !all { return err }
						errors = append(errors, err)
					}
				{{- end }}
			{{- else if $r.Gt }}
				if dur <= gt {
					err := {{GetAliasName}}ValidationError(
						{{.ErrCode}}, 
						"[{{ $f.GoName }}] value must be greater than {{durStr $r.GetGt}}",
						nil,
					)
					if !all { return err }
					errors = append(errors, err)
				}
			{{- else if $r.Gte }}
				if dur < gte {
					err := {{GetAliasName}}ValidationError(
						{{.ErrCode}}, 
						"[{{ $f.GoName }}] value must be greater than or equal to {{durStr $r.GetGte}}",
						nil,
					)
					if !all { return err }
					errors = append(errors, err)
				}
			{{- end }}


			{{- if $r.In }}
				if _, ok := {{ lookup $f "InLookup" }}[dur]; !ok {
					err := {{GetAliasName}}ValidationError(
						{{.ErrCode}}, 
						"[{{ $f.GoName }}] value must be in {{ $r.In }}",
						nil,
					)
					if !all { return err }
					errors = append(errors, err)
				}
			{{- else if $r.NotIn }}
				if _, ok := {{ lookup $f "NotInLookup" }}[dur]; ok {
					err := {{GetAliasName}}ValidationError(
						{{.ErrCode}}, 
						"[{{ $f.GoName }}] value must not be in {{ $r.NotIn }}",
						nil,
					)
					if !all { return err }
					errors = append(errors, err)
				}
			{{ end }}
`
