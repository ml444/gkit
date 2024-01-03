package templates

const TimestampTpl = `{{ $f := .Field }}{{ $r := .Rules }}
	{{ template "required" . }}

	{{ if or $r.Lt $r.Lte $r.Gt $r.Gte $r.LtNow $r.GtNow $r.Within $r.Const }}
		if t := {{ accessor . }}; t != nil {
			ts, err := t.AsTime(), t.CheckValid()
			if err != nil {
				err = {{ errCause .Field "err" "value is not a valid timestamp" }}
				if !all { return err }
				errors = append(errors, err)
			} else {
				{{ template "timestampcmp" . }}
			}
		}
	{{ end }}
`
const TimestampcmpTpl = `{{ $f := .Field }}{{ $r := .Rules }}
			{{  if $r.Const }}
				if !ts.Equal({{ tsLit $r.Const }}) {
					err := {{ err .Field "value must equal " (tsStr $r.Const) }}
					if !all { return err }
					errors = append(errors, err)
				}
			{{ end }}

			{{ if or $r.LtNow $r.GtNow $r.Within }} now := time.Now(); {{ end }}
			{{- if $r.Lt }}  lt  := {{ tsLit $r.Lt }};  {{ end }}
			{{- if $r.Lte }} lte := {{ tsLit $r.Lte }}; {{ end }}
			{{- if $r.Gt }}  gt  := {{ tsLit $r.Gt }};  {{ end }}
			{{- if $r.Gte }} gte := {{ tsLit $r.Gte }}; {{ end }}
			{{- if $r.Within }} within := {{ durLit $r.Within }}; {{ end }}

			{{ if $r.Lt }}
				{{ if $r.Gt }}
					{{  if tsGt $r.GetLt $r.GetGt }}
						if ts.Sub(gt) <= 0 || ts.Sub(lt) >= 0 {
							err := {{ err .Field "value must be inside range (" (tsStr $r.GetGt) ", " (tsStr $r.GetLt) ")" }}
							if !all { return err }
							errors = append(errors, err)
						}
					{{ else }}
						if ts.Sub(lt) >= 0 && ts.Sub(gt) <= 0 {
							err := {{ err .Field "value must be outside range [" (tsStr $r.GetLt) ", " (tsStr $r.GetGt) "]" }}
							if !all { return err }
							errors = append(errors, err)
						}
					{{ end }}
				{{ else if $r.Gte }}
					{{  if tsGt $r.GetLt $r.GetGte }}
						if ts.Sub(gte) < 0 || ts.Sub(lt) >= 0 {
							err := {{ err .Field "value must be inside range [" (tsStr $r.GetGte) ", " (tsStr $r.GetLt) ")" }}
							if !all { return err }
							errors = append(errors, err)
						}
					{{ else }}
						if ts.Sub(lt) >= 0 && ts.Sub(gte) < 0 {
							err := {{ err .Field "value must be outside range [" (tsStr $r.GetLt) ", " (tsStr $r.GetGte) ")" }}
							if !all { return err }
							errors = append(errors, err)
						}
					{{ end }}
				{{ else }}
					if ts.Sub(lt) >= 0 {
						err := {{ err .Field "value must be less than " (tsStr $r.GetLt) }}
						if !all { return err }
						errors = append(errors, err)
					}
				{{ end }}
			{{ else if $r.Lte }}
				{{ if $r.Gt }}
					{{  if tsGt $r.GetLte $r.GetGt }}
						if ts.Sub(gt) <= 0 || ts.Sub(lte) > 0 {
							err := {{ err .Field "value must be inside range (" (tsStr $r.GetGt) ", " (tsStr $r.GetLte) "]" }}
							if !all { return err }
							errors = append(errors, err)
						}
					{{ else }}
						if ts.Sub(lte) > 0 && ts.Sub(gt) <= 0 {
							err := {{ err .Field "value must be outside range (" (tsStr $r.GetLte) ", " (tsStr $r.GetGt) "]" }}
							if !all { return err }
							errors = append(errors, err)
						}
					{{ end }}
				{{ else if $r.Gte }}
					{{ if tsGt $r.GetLte $r.GetGte }}
						if ts.Sub(gte) < 0 || ts.Sub(lte) > 0 {
							err := {{ err .Field "value must be inside range [" (tsStr $r.GetGte) ", " (tsStr $r.GetLte) "]" }}
							if !all { return err }
							errors = append(errors, err)
						}
					{{ else }}
						if ts.Sub(lte) > 0 && ts.Sub(gte) < 0 {
							err := {{ err .Field "value must be outside range (" (tsStr $r.GetLte) ", " (tsStr $r.GetGte) ")" }}
							if !all { return err }
							errors = append(errors, err)
						}
					{{ end }}
				{{ else }}
					if ts.Sub(lte) > 0 {
						err := {{ err .Field "value must be less than or equal to " (tsStr $r.GetLte) }}
						if !all { return err }
						errors = append(errors, err)
					}
				{{ end }}
			{{ else if $r.Gt }}
				if ts.Sub(gt) <= 0 {
					err := {{ err .Field "value must be greater than " (tsStr $r.GetGt) }}
					if !all { return err }
					errors = append(errors, err)
				}
			{{ else if $r.Gte }}
				if ts.Sub(gte) < 0 {
					err := {{ err .Field "value must be greater than or equal to " (tsStr $r.GetGte) }}
					if !all { return err }
					errors = append(errors, err)
				}
			{{ else if $r.LtNow }}
				{{ if $r.Within }}
					if ts.Sub(now) >= 0 || ts.Sub(now.Add(-within)) < 0 {
						err := {{ err .Field "value must be less than now within " (durStr $r.GetWithin) }}
						if !all { return err }
						errors = append(errors, err)
					}
				{{ else }}
					if ts.Sub(now) >= 0 {
						err := {{ err .Field "value must be less than now" }}
						if !all { return err }
						errors = append(errors, err)
					}
				{{ end }}
			{{ else if $r.GtNow }}
				{{ if $r.Within }}
					if ts.Sub(now) <= 0 || ts.Sub(now.Add(within)) > 0 {
						err := {{ err .Field "value must be greater than now within " (durStr $r.GetWithin) }}
						if !all { return err }
						errors = append(errors, err)
					}
				{{ else }}
					if ts.Sub(now) <= 0 {
						err := {{ err .Field "value must be greater than now" }}
						if !all { return err }
						errors = append(errors, err)
					}
				{{ end }}
			{{ else if $r.Within }}
				if ts.Sub(now.Add(within)) >= 0 || ts.Sub(now.Add(-within)) <= 0 {
					err := {{ err .Field "value must be within " (durStr $r.GetWithin) " of now" }}
					if !all { return err }
					errors = append(errors, err)
				}
			{{ end }}
`
