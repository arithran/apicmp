package diff

import "text/template"

const curlTemplate = `
Testing Row: {{.Row}}
===============
Before:
curl --location --request {{ .Before.Method }} '{{ .Before.Path }}' \{{range $k, $v := .Before.Headers}}
--header '{{$k}}: {{$v}}'{{end}}

After:
curl --location --request {{ .After.Method }} '{{ .After.Path }}' \{{range $k, $v := .After.Headers}}
--header '{{$k}}: {{$v}}'{{end}}

Result:
`

var curlTpl = template.Must(template.New("curl").Parse(curlTemplate))
