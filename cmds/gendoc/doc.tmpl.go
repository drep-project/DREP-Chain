package main

const doctmpl = `
#  `+"`"+`JSON-RPC`+"`"+`接口说明文档
## {{range .}}{{ .StructDocStr | trim | html}}
{{ .FuncDocStr | html}}{{end}}`

const methodtmpl = `{{range $index2, $method := .}}
### {{inc $index2}}  . {{$method.Prefix}}_{{$method.Name}}
#### 作用：{{ with (index .Tokens "usage:") }}{{ .Str }}{{ end }}
> 参数：
{{ with (index $method.Tokens "params:") }}{{range $index, $params := .Params}} {{inc $index}}. {{$params}}
{{end}}{{ end }}
#### 返回值：{{ with (index $method.Tokens "return:") }}{{.Str | html}}{{ end }}

#### 示例代码
##### 请求：

`+"```"+`shell
{{ with (index $method.Tokens "example:") }}{{.Str | html}}{{ end }}
`+"```"+`

##### 响应：

`+"```"+`json
{{ with (index $method.Tokens "response:") }}{{ .Str | html }}{{ end }}
`+"````" +`

{{end}}`


const structtmpl = `
{{.Name}}
{{ with (index .Tokens "usage:") }}{{ .Str }}{{ end }}`
