package main

const doctmpl = `
#  ` + "`" + `JSON-RPC` + "`" + `RPC interface
## {{range .}}{{ .StructDocStr | trim | html}}
{{ .FuncDocStr | html}}{{end}}`

const methodtmpl = `{{range $index2, $method := .}}
### {{inc $index2}}. {{$method.Prefix}}_{{$method.Name}}
#### usage：{{ with (index .Tokens "usage:") }}{{ .Str }}{{ end }}
> params：
{{ with (index $method.Tokens "params:") }}{{range $index, $params := .Params}} {{inc $index}}. {{$params}}
{{end}}{{ end }}
#### return：{{ with (index $method.Tokens "return:") }}{{.Str | html}}{{ end }}

#### example

` + "```" + `shell
{{ with (index $method.Tokens "example:") }}{{.Str | html}}{{ end }}
` + "```" + `

##### response：

` + "```" + `json
{{ with (index $method.Tokens "response:") }}{{ .Str | html }}{{ end }}
` + "````" + `

{{end}}`

const structtmpl = `
{{.Name}}
{{ with (index .Tokens "usage:") }}{{ .Str }}{{ end }}`
