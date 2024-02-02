package templates

const LtGtTpl = `{{ $f := .Field }}{{ $r := .Rules }}
	{{ if $r.Lt }}
		{{ if $r.Gt }}
			{{  if gt $r.GetLt $r.GetGt }}
				if val := {{ .GetAccessor }};  val <= {{ $r.Gt }} || val >= {{ $r.Lt }} {
					err := {{ err .Field "value must be inside range (" $r.GetGt ", " $r.GetLt ")" }}
					if !all { return err }
					errors = append(errors, err)
				}
			{{ else }}
				if val := {{ .GetAccessor }}; val >= {{ $r.Lt }} && val <= {{ $r.Gt }} {
					err := {{ err .Field "value must be outside range [" $r.GetLt ", " $r.GetGt "]" }}
					if !all { return err }
					errors = append(errors, err)
				}
			{{ end }}
		{{ else if $r.Gte }}
			{{  if gt $r.GetLt $r.GetGte }}
				if val := {{ .GetAccessor }};  val < {{ $r.Gte }} || val >= {{ $r.Lt }} {
					err := {{ err .Field "value must be inside range [" $r.GetGte ", " $r.GetLt ")" }}
					if !all { return err }
					errors = append(errors, err)
				}
			{{ else }}
				if val := {{ .GetAccessor }}; val >= {{ $r.Lt }} && val < {{ $r.Gte }} {
					err := {{ err .Field "value must be outside range [" $r.GetLt ", " $r.GetGte ")" }}
					if !all { return err }
					errors = append(errors, err)
				}
			{{ end }}
		{{ else }}
			if {{ .GetAccessor }} >= {{ $r.Lt }} {
				err := {{ err .Field "value must be less than " $r.GetLt }}
				if !all { return err }
				errors = append(errors, err)
			}
		{{ end }}
	{{ else if $r.Lte }}
		{{ if $r.Gt }}
			{{  if gt $r.GetLte $r.GetGt }}
				if val := {{ .GetAccessor }};  val <= {{ $r.Gt }} || val > {{ $r.Lte }} {
					err := {{ err .Field "value must be inside range (" $r.GetGt ", " $r.GetLte "]" }}
					if !all { return err }
					errors = append(errors, err)
				}
			{{ else }}
				if val := {{ .GetAccessor }}; val > {{ $r.Lte }} && val <= {{ $r.Gt }} {
					err := {{ err .Field "value must be outside range (" $r.GetLte ", " $r.GetGt "]" }}
					if !all { return err }
					errors = append(errors, err)
				}
			{{ end }}
		{{ else if $r.Gte }}
			{{ if gt $r.GetLte $r.GetGte }}
				if val := {{ .GetAccessor }};  val < {{ $r.Gte }} || val > {{ $r.Lte }} {
					err := {{ err .Field "value must be inside range [" $r.GetGte ", " $r.GetLte "]" }}
					if !all { return err }
					errors = append(errors, err)
				}
			{{ else }}
				if val := {{ .GetAccessor }}; val > {{ $r.Lte }} && val < {{ $r.Gte }} {
					err := {{ err .Field "value must be outside range (" $r.GetLte ", " $r.GetGte ")" }}
					if !all { return err }
					errors = append(errors, err)
				}
			{{ end }}
		{{ else }}
			if {{ .GetAccessor }} > {{ $r.Lte }} {
				err := {{ err .Field "value must be less than or equal to " $r.GetLte }}
				if !all { return err }
				errors = append(errors, err)
			}
		{{ end }}
	{{ else if $r.Gt }}
		if {{ .GetAccessor }} <= {{ $r.Gt }} {
			err := {{ err .Field "value must be greater than " $r.GetGt }}
			if !all { return err }
			errors = append(errors, err)
		}
	{{ else if $r.Gte }}
		if {{ .GetAccessor }} < {{ $r.Gte }} {
			err := {{ err .Field "value must be greater than or equal to " $r.GetGte }}
			if !all { return err }
			errors = append(errors, err)
		}
	{{ end }}
`
