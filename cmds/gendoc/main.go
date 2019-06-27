package main

import (
	"bytes"
	"html/template"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"strings"
)

const (
	MidPath = "src/github.com/drep-project/drep-chain/"
)

var (
	structTemplate *template.Template
	methodTemplate *template.Template
	docTemplate    *template.Template
)

func main() {
	//detect which gopath is worked
	gopath, err := detectSource()
	if err != nil {
		log.Fatal(err)
		return
	}
	drepChainPath := filepath.Join(gopath, MidPath)
	apiPackages := getAllPackage(drepChainPath)

	//parser all package,
	allPackages := []*PackageParser{}
	for _, apiPackage := range apiPackages {
		fullPath := filepath.Join(gopath, MidPath, apiPackage)
		p := NewPackageParser(fullPath)
		p.parserPackage()
		allPackages = append(allPackages, p)
	}

	docs := []*RpcDoc{}
	for _, packageParser := range allPackages {
		for _, define := range packageParser.Defines {
			lowcaseName := strings.ToLower(define.Name)
			if !strings.HasSuffix(lowcaseName, "api") || len(lowcaseName) <= 3 {
				continue
			}

			rpcDoc := NewRpcDoc()
			structDocStr := ExtractDoc(define.Comment)
			if structDocStr == "" {
				continue
			}
			rpcDoc.StructDoc = structParser(structDocStr)

			if val, ok := rpcDoc.StructDoc.Tokens[NAME]; ok {
				if val.Str == "" {
					continue
				}
			} else {
				continue
			}
			for _, funcDeifine := range define.Methods {
				funcCommentStr := ExtractDoc(funcDeifine.Doc)
				if funcCommentStr == "" {
					continue
				}
				funcCommentDoc := funcParser(funcCommentStr, rpcDoc.StructDoc.Tokens[PREFIX].Str)
				rpcDoc.FuncDoc = append(rpcDoc.FuncDoc, funcCommentDoc)
			}

			//struct
			structBuffer := bytes.NewBuffer([]byte{})
			err := structTemplate.Execute(structBuffer, rpcDoc.StructDoc)
			if err != nil {
				log.Fatal(err)
				return
			}
			rpcDoc.StructDocStr = string(structBuffer.Bytes())
			//method
			methodBuffer := bytes.NewBuffer([]byte{})
			err = methodTemplate.Execute(methodBuffer, rpcDoc.FuncDoc)
			if err != nil {
				log.Fatal(err)
				return
			}
			rpcDoc.FuncDocStr = string(methodBuffer.Bytes())
			docs = append(docs, rpcDoc)
		}
	}

	//method
	docBuffer := bytes.NewBuffer([]byte{})
	err = docTemplate.Execute(docBuffer, docs)
	if err != nil {
		log.Fatal(err)
		return
	}
	docPath := filepath.Join(drepChainPath, "doc", "JSON-RPC.md")
	err = ioutil.WriteFile(docPath, docBuffer.Bytes(), os.ModePerm)
	if err != nil {
		log.Fatal(err)
		return
	}
}

func init() {
	structTemplate = template.New("struct.tmpl.go")
	docTemplate = template.New("doc.tmpl.go")
	methodTemplate = template.New("method.tmpl.go")

	funcs := map[string]interface{}{}
	funcs["html"] = func(x string) interface{} {
		return template.HTML(x)
	}
	funcs["inc"] = func(i int) int {
		return i + 1
	}
	funcs["trim"] = func(str string) string {
		return strings.Trim(str, "\n\r ")
	}
	structTemplate.Funcs(funcs)
	methodTemplate.Funcs(funcs)
	docTemplate.Funcs(funcs)

	_, err := structTemplate.Parse(structtmpl)
	if err != nil {
		log.Fatal(err)
	}

	_, err = methodTemplate.Parse(methodtmpl)
	if err != nil {
		log.Fatal(err)
	}

	_, err = docTemplate.Parse(doctmpl)
	if err != nil {
		log.Fatal(err)
	}
}

type RpcDoc struct {
	StructDoc *StructDoc
	FuncDoc   []*FuncDoc

	StructDocStr string
	FuncDocStr   string
}

func NewRpcDoc() *RpcDoc {
	return &RpcDoc{
		FuncDoc: []*FuncDoc{},
	}
}
