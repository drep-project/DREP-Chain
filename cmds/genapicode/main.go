package main

import (
	"fmt"
	"github.com/drep-project/DREP-Chain/blockmgr"
	chainService "github.com/drep-project/DREP-Chain/chain"
	p2pService "github.com/drep-project/DREP-Chain/network/service"
	accountService "github.com/drep-project/DREP-Chain/pkgs/accounts/service"
	logService "github.com/drep-project/DREP-Chain/pkgs/log"
	traceService "github.com/drep-project/DREP-Chain/pkgs/trace"
	"io"
	"os"
	path2 "path"
	"reflect"
	"strings"
)

const codeFile = `
var Method = require('../method');
var formatters = require('../formatters');
var utils = require('../../utils/utils');

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

var (
	inputFormatMap = map[string]string{
		"common.Big":           "utils.fromDecimal",
		"int64":                "utils.fromDecimal",
		"uint64":               "utils.fromDecimal",
		"crypto.CommonAddress": "formatters.inputAddressFormatter",
	}
	formatMap = map[string]string{
		"big.Int": "formatters.outputBigNumberFormatter",
		"int64":   "utils.toDecimal",
		"MeInfo":  "formatters.meInfoFormatter",
		"Storage": "formatters.storageFormatter",
	}
)

func Capitalize(str string) string {
	var upperStr string
	vv := []rune(str)
	for i := 0; i < len(vv); i++ {
		if i == 0 {
			if vv[i] >= 97-32 && vv[i] <= 122-32 {
				vv[i] += 32
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

	output := "std"
	if len(os.Args) > 0 {
		output = "file"
	}

	vType := reflect.TypeOf(&p2pService.P2PApi{})
	resolveType(output, "p2p", "P2P", "p2p", vType)

	vType = reflect.TypeOf(&accountService.AccountApi{})
	resolveType(output, "account", "ACCOUNT", "account", vType)

	vType = reflect.TypeOf(&blockmgr.BlockMgrApi{})
	resolveType(output, "blockmgr", "BLOCKMGR", "blockmgr", vType)

	vType = reflect.TypeOf(&logService.LogApi{})
	resolveType(output, "log", "LOG", "log", vType)

	vType = reflect.TypeOf(&chainService.ChainApi{})
	resolveType(output, "chain", "CHAIN", "chain", vType)

	//vType = reflect.TypeOf(&solo.ConsensusApi{})
	//resolveType(output, "consensus", "CONSENSUS", "consensus", vType)

	vType = reflect.TypeOf(&traceService.TraceApi{})
	resolveType(output, "trace", "TRACE", "trace", vType)
}

func resolveType(output string, fileName, className string, prefix string, vType reflect.Type) {
	fmt.Println("**********" + fileName + "***************")
	code := generateCode(className, prefix, vType)
	if output == "std" {
		fmt.Println(code)
	} else {
		WriteFile(fileName+".js", code)
	}
}
func generateCode(className string, prefix string, vType reflect.Type) string {
	methods := vType.NumMethod()

	template := `
var %s = new Method({
	name: '%s',
	call: '%s_%s',
	params: %d,%s
    inputFormatter: [%s]
});
	`
	code := ""
	methodNames := ""

	for i := 0; i < methods; i++ {
		m := vType.Method(i)
		numIn := m.Func.Type().NumIn()
		oNmae := m.Name
		methodName := Capitalize(oNmae)
		name := ""
		if m.Func.Type().NumOut() > 0 {
			var resultType = m.Func.Type().Out(0)
			if resultType.Kind() == reflect.Ptr {
				resultType = resultType.Elem()
				name = resultType.Name()
			} else {
				name = resultType.Name()
			}
		}
		outputFormater := ""
		if formater, ok := formatMap[name]; ok {
			outputFormater = "\n\toutputFormatter : " + formater
		}

		inputFormat := ""
		if m.Func.Type().NumIn() > 1 {
			for j := 1; j < m.Func.Type().NumIn(); j++ {
				var inputType = m.Func.Type().In(j)
				inputTypeName := ""
				if inputType.Kind() == reflect.Ptr {
					inputType = inputType.Elem()
					inputTypeName = inputType.String()
				} else {
					inputTypeName = inputType.String()
				}
				if formater, ok := inputFormatMap[inputTypeName]; ok {
					inputFormat = inputFormat + formater + ", "
				} else {
					inputFormat = inputFormat + "null, "
				}
			}
			if len(inputFormat) > 0 {
				inputFormat = inputFormat[:len(inputFormat)-2]
			}
		}

		code += fmt.Sprintf(template, methodName, methodName, prefix, methodName, numIn-1, outputFormater, inputFormat)
		methodNames += methodName + ","
	}
	methodNames = strings.Trim(methodNames, ",")

	codestr := fmt.Sprintf(codeFile, className, code, methodNames, className)
	return codestr
}

func WriteFile(name, content string) {
	rootDir := getCurPath()
	path := path2.Join(rootDir, name)
	fileObj, err := os.OpenFile(path, os.O_RDWR|os.O_CREATE, 0644)
	if err != nil {
		fmt.Println("Failed to open the file", err.Error())
		os.Exit(2)
	}
	if _, err := io.WriteString(fileObj, content); err == nil {
		fmt.Println("Successful appending to the file with os.OpenFile and io.WriteString.")
	}
}

func getCurPath() string {
	dir, _ := os.Getwd()
	return dir
}
