package templates

const NumTpl = `
	{{ if .Rules.GetIgnoreEmpty }}
		if {{ .GetAccessor }} != 0 {
	{{ end }}

	{{ template "const" . }}
	{{ template "ltgt" . }}
	{{ template "in" . }}

	{{ if .Rules.GetIgnoreEmpty }}
		}
	{{ end }}

`
