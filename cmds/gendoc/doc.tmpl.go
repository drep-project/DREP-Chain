package main

const doctmpl = `
#  RPC interface
## {{range .}}{{ .StructDocStr | trim | html}}
{{ .FuncDocStr | html}}{{end}}`

const methodtmpl = `{{range $index2, $method := .}}
## {{inc $index2}}. {{$method.Prefix}}_{{$method.Name}}
<font size=5> Usage：{{ with (index .Tokens "usage:") }}{{ .Str }}{{ end }} </font>  

<font size=5> Params：</font>  

{{ with (index $method.Tokens "params:") }}{{range $index, $params := .Params}} {{inc $index}}. {{$params}}
{{end}}{{ end }}
<font size=5> Return：{{ with (index $method.Tokens "return:") }}{{.Str | html}}{{ end }} </font>

<font size=5> Example: </font>   

<font size=4> shell: </font>
` + "```" + `shell
{{ with (index $method.Tokens "example:") }}{{.Str | html}}{{ end }}
` + "```" + `
<font size=4> cli: </font>
` + "```" + `cli
drepClient 127.0.0.1:10085 {{$method.Prefix}}_{{$method.Name}} 3
` + "```" + `

<font size=5> Response：</font>

` + "```" + `json
{{ with (index $method.Tokens "response:") }}{{ .Str | html }}{{ end }}
` + "```" + `

{{end}}`

const structtmpl = `
{{.Name}}
{{ with (index .Tokens "usage:") }}{{ .Str }}{{ end }}`
