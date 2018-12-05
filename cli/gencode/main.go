package main 

import(
	"fmt"
	"reflect"
	"BlockChainTest/node"
)


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
	vType:=reflect.TypeOf(&node.ChainApi{})

    methods := vType.NumMethod()


	template := `
var %s = new Method({
	name: '%s',
	call: 'db_%s',
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
		
		code += fmt.Sprintf(template,methodName, methodName, methodName,numIn)
		methodNames += methodName +","
	}

	fmt.Println(code)
	fmt.Println(methodNames)
}