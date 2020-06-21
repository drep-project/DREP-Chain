package genDrepClientCode

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"strings"
)

type callStruct struct {
	Jsonrpc string        `json:"jsonrpc"`
	Method  string        `json:"method"`
	Params  []interface{} `json:"params,omitempty"`
	ID      int           `json:"id"`
}

func main() {
	// 生成代码
	genInitFunc()
	genMethodFunc()
}

func genInitFunc() {
	//打开文件
	file, err := os.Open("./doc/JSON-RPC.md")
	if err != nil {
		fmt.Println("open json-rpc.md err:", err)
		return
	}

	defer file.Close()

	i := 0

	fmt.Println("func init() {")

	// 逐行遍历
	br := bufio.NewReader(file)
	for {
		s, _, c := br.ReadLine()
		if c == io.EOF {
			break
		}

		line := string(s)

		if strings.Contains(line, "curl") {

			beginJson := strings.Index(line, "{")
			endJson := strings.Index(line, "}")
			line = line[beginJson : endJson+1]

			c := callStruct{}
			err := json.Unmarshal([]byte(line), &c)
			if err != nil {
				fmt.Println("//json unmarshal err:", err, "line:", line)
				//return
				continue
			}

			methodName := strings.Split(c.Method, "_")
			fmt.Printf("	methods[\"%s\"] = %s\n", c.Method, methodName[1])
			i++
		}
	}
	fmt.Println("}")

	//fmt.Println("method number :", i)
}

func genMethodFunc() {
	//打开文件
	file, err := os.Open("./doc/JSON-RPC.md")
	if err != nil {
		fmt.Println("open json-rpc.md err:", err)
		return
	}

	methodNum := 0
	defer file.Close()
	// 逐行遍历
	br := bufio.NewReader(file)
	for {
		s, _, c := br.ReadLine()
		if c == io.EOF {
			break
		}

		line := string(s)

		//curl http://localhost:10085 -X POST --data
		// '{"jsonrpc":"2.0","method":"blockmgr_getTransactionCount","params":["0x8a8e541ddd1272d53729164c70197221a3c27486"], "id": 3}'
		// -H "Content-Type:application/json"
		if strings.Contains(line, "curl") {
			beginJson := strings.Index(line, "{")
			endJson := strings.Index(line, "}")
			line = line[beginJson : endJson+1]
			c := callStruct{}
			err := json.Unmarshal([]byte(line), &c)
			if err != nil {
				fmt.Println("//json unmarshal err:", err, line)
				fmt.Println()
				//return
				continue
			}

			methodName := strings.Split(c.Method, "_")
			fmt.Printf("func %s(args cli.Args, client *rpc.Client, ctx context.Context) {\n", methodName[1])
			fmt.Printf("	var resp interface\n")

			callLine := "	if err := client.CallContext(ctx, &resp"

			for i := 0; i < len(c.Params)+1; i++ {
				callLine += fmt.Sprintf(", args[%d]", i)
			}

			callLine += "); err != nil {"
			fmt.Println(callLine)

			fmt.Println("		fmt.Println(err)")
			fmt.Println("	}")
			fmt.Println("	fmt.Println(resp)")

			fmt.Println("}")
			methodNum++
		}
	}

	//fmt.Println("method num:", methodNum)
}
