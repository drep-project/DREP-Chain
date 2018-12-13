package main 

import(
	"fmt"
	"os"
	"io"
	"reflect"
	"strings"
	"BlockChainTest/accounts"
	"BlockChainTest/database"
	"BlockChainTest/node"
)

const codeFile = 
`
var Method = require('../method');

var %s = function (drep) {
    this._requestManager = drep._requestManager;

    var self = this;
    
    methods().forEach(function(method) { 
        method.attachToObject(self);
        method.setRequestManager(drep._requestManager);
    });
};

var methods = function () {
	%s
    return [%s]
}

module.exports = %s;
`

func Capitalize(str string) string {
    var upperStr string
    vv := []rune(str)   // 后文有介绍
    for i := 0; i < len(vv); i++ {
        if i == 0 {
            if vv[i] >= 97-32 && vv[i] <= 122-32 {  // 后文有介绍
                vv[i] += 32 // string的码表相差32位
                upperStr += string(vv[i])
            } else {
                fmt.Println("Not begins with lowercase letter,")
                return str
            }
        } else {
            upperStr += string(vv[i])
        }
    }
    return upperStr
}

func main() {

	output  := "std"
	if len(os.Args) >0 {
		if os.Args[1] == "file" {
			output = "file"
		}
	}

	vType:=reflect.TypeOf(&database.DataBaseAPI{})
	resolveType(output,"db", "DB", "db",vType)

	vType=reflect.TypeOf(&accounts.AccountApi{})
	resolveType(output,"account", "ACCOUNT", "account",vType)

	vType=reflect.TypeOf(&node.ChainApi{})
	resolveType(output,"chain", "CHAIN", "chain",vType)
}

func resolveType(output string,fileName, className string,prefix string, vType reflect.Type){
	fmt.Println("**********"+fileName+"***************")
	vType=reflect.TypeOf(&accounts.AccountApi{})
	code := generateCode(className, prefix,vType)
	if output == "std" {
		fmt.Println(code)
	}else{
		WriteFile(fileName+".js",code)
	}
}
func generateCode(className string,prefix string, vType reflect.Type) string{
	methods := vType.NumMethod()


	template := `
var %s = new Method({
	name: '%s',
	call: '%s_%s',
	params: %d
});
	`
	code := ""
	methodNames := ""
	for i:= 0 ;i < methods;i++{
		m := vType.Method(i)
		numIn:=m.Func.Type().NumIn()
		oNmae := m.Name
		methodName := Capitalize(oNmae)
		
		code += fmt.Sprintf(template,methodName, methodName,prefix, methodName,numIn-1)
		methodNames += methodName +","
	}
	methodNames = strings.Trim(methodNames,",")

	codestr := fmt.Sprintf(codeFile, className, code, methodNames, className)
	return codestr
}

func WriteFile(name,content string) {
    fileObj,err := os.OpenFile(name,os.O_RDWR|os.O_CREATE,0644)
    if err != nil {
        fmt.Println("Failed to open the file",err.Error())
        os.Exit(2)
    }
    if  _,err := io.WriteString(fileObj,content);err == nil {
        fmt.Println("Successful appending to the file with os.OpenFile and io.WriteString.",content)
    }
}