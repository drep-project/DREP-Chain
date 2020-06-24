package main

const doctmpl = `
#  RPC interface
## {{range .}}{{ .StructDocStr | trim | html}}
{{ .FuncDocStr | html}}{{end}}`

const methodtmpl = `{{range $index2, $method := .}}
### {{inc $index2}}. {{$method.Prefix}}_{{$method.Name}}
- **Usage：**  

&emsp;&emsp;{{ with (index .Tokens "usage:") }}{{ .Str }}{{ end }}

- **Params：**  

&emsp;&emsp;{{ with (index $method.Tokens "params:") }}{{range $index, $params := .Params}} {{inc $index}}. {{$params}}
{{end}}{{ end }}
- **Return：{{ with (index $method.Tokens "return:") }}{{.Str | html}}{{ end }}**

- **Example:**  

**shell:**
` + "```" + `shell
{{ with (index $method.Tokens "example:") }}{{.Str | html}}{{ end }}
` + "```" + `
**cli:**
` + "```" + `cli
drepClient 127.0.0.1:10085 {{$method.Prefix}}_{{$method.Name}} 3
` + "```" + `

- **Response：**

` + "```" + `json
{{ with (index $method.Tokens "response:") }}{{ .Str | html }}{{ end }}
` + "```" + `

---

{{end}}`

const structtmpl = `
{{.Name}}
{{ with (index .Tokens "usage:") }}{{ .Str }}{{ end }}`
