package diff

import "text/template"

const curlTemplate = `
Testing Row: {{.Row}}
===============
Before:
curl --location --request {{ .Before.Method }} '{{ .Before.Path }}' \{{range $k, $v := .Before.Headers}}
--header '{{$k}}: {{$v}}' \{{end}}{{if ne (len .Before.Body) 0}}
--data-raw '{{ .Before.Body }}'{{end}}

After:
curl --location --request {{ .After.Method }} '{{ .After.Path }}' \{{range $k, $v := .After.Headers}}
--header '{{$k}}: {{$v}}' \{{end}}{{if ne (len .After.Body) 0}}
--data-raw '{{ .After.Body }}'{{end}}

`

const summaryTemplate = `
Summary:
  Total Tests : {{.Count}}
  Passed      : {{.Passed}}
  Failed      : {{.Failed}}
  Failed Rows : {{.FailedRowsStr}}
  Time        : {{.Time}}

Issues Found:
`

var tpl *template.Template

func init() {
	tpl = template.Must(template.New("curl").Parse(curlTemplate))
	tpl = template.Must(tpl.New("summary").Parse(summaryTemplate))
}
