package main

import (
	"go/ast"
	"fmt"
	"go/parser"
	"go/token"
	"log"
	"strings"
)

type PackageParser struct {
	Dir      string
	Packages map[string]*ast.Package
	Defines  map[string]*Define
}

type Define struct {
	Name    string
	Comment *ast.CommentGroup
	Spec    *ast.TypeSpec
	Methods []*ast.FuncDecl
}

func NewPackageParser(dir string) *PackageParser {
	fset := token.NewFileSet()
	parserMode := parser.ParseComments
	packages, err := parser.ParseDir(fset, dir, nil, parserMode)
	if err != nil {
		log.Fatal(err)
	}
	return &PackageParser{
		Dir:      dir,
		Packages: packages,
		Defines:  make(map[string]*Define),
	}
}

func (packageParser *PackageParser) scanTypes() {

}

func (packageParser *PackageParser) parserPackage() {
	for _, packagee := range packageParser.Packages {
		for _, file := range packagee.Files {
			for _, decl := range file.Decls {
				switch decl := decl.(type) {
				case *ast.GenDecl:
					packageParser.parserStuctSpect(decl)
				case *ast.FuncDecl:
					if isStructMethod(decl) && isMethodPublic(decl) {
						packageParser.parserFuncRecv(decl)
					}
				}
			}
		}
	}

}

func (packageParser *PackageParser) parserFuncRecv(decl *ast.FuncDecl) {
	ref := decl.Recv.List[0]
	var refTypeName string
	var refType *ast.TypeSpec
	switch ref := ref.Type.(type) {
	case *ast.StarExpr:
		switch ref := ref.X.(type) {
		case *ast.Ident:
			if ref.Obj == nil {
				refTypeName = ref.Name
			} else {
				switch ref := ref.Obj.Decl.(type) {
				case *ast.TypeSpec:
					refType = ref
					refTypeName = ref.Name.Name
				}
			}
		}
	case *ast.Ident:
		if ref.Obj == nil {
			refTypeName = ref.Name
		} else {
			switch ref := ref.Obj.Decl.(type) {
			case *ast.TypeSpec:
				refType = ref
				refTypeName = ref.Name.Name
			}
		}
	}

	define, ok := packageParser.Defines[refTypeName]
	if ok {
		define.Methods = append(define.Methods, decl)
	} else {
		packageParser.Defines[refTypeName] = &Define{
			Spec:    refType,
			Name:    refTypeName,
			Methods: []*ast.FuncDecl{decl},
		}
	}
}

func (packageParser *PackageParser) parserStuctSpect(genDecl *ast.GenDecl) {
	for _, spec := range genDecl.Specs {
		switch spec := spec.(type) {
		case *ast.TypeSpec:
			val, ok := packageParser.Defines[spec.Name.Name]
			if ok {
				if val.Spec == nil {
					val.Spec = spec
					val.Comment = genDecl.Doc
				}
				return
			}
			packageParser.Defines[spec.Name.Name] = &Define{
				Spec:    spec,
				Name:    spec.Name.Name,
				Comment : genDecl.Doc,
				Methods: []*ast.FuncDecl{},
			}
		default:
			log.Println("skip no type spec")
		}
	}
}

//
func ExtractDoc(comments *ast.CommentGroup) string {
	comment := ""
	if comments ==nil || comments.List == nil {
		return ""
	}
	for _, doc := range comments.List {
		if strings.HasPrefix(doc.Text, "//") {
			comment = comment + " " + strings.TrimLeft(doc.Text, "\\")
		}else{
			text := strings.TrimLeft(doc.Text, "/*")
			text = strings.TrimRight(text, "*/")
			comment = comment + " " + text
		}

	}
	return comment

	/**/
}

func isStructMethod(method *ast.FuncDecl) bool {
	if method.Name.Obj == nil {
		return true
	}
	addr1 := fmt.Sprintf("%p\n", method.Name.Obj.Decl)
	addr2 := fmt.Sprintf("%p\n", method)
	return addr1 != addr2
}

func isMethodPublic(method *ast.FuncDecl) bool {
	ch := method.Name.Name[0]
	return ch >= 'A' && ch <= 'Z'
}
