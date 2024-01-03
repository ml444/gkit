package templates

const NoneTpl = `// no validation rules for {{ .Field.GoName }}
	{{- if .Index }}[{{ .Index }}]{{ end }}
	{{- if .OnKey }} (key){{ end }}`
