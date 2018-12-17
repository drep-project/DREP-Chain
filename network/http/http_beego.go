package http

import (
    "github.com/astaxie/beego"
    "flag"
    "fmt"
    "strconv"
    "BlockChainTest/database"
    "BlockChainTest/bean"
    "encoding/json"
    "math/big"
    "BlockChainTest/node"
    "BlockChainTest/accounts"
    "BlockChainTest/config"
)

var mappingMethodMap = map[string] string {
    "/GetAllBlocks": "*:GetAllBlocks",
    "/GetBlock": "*:GetBlock",
    "/GetHighestBlock": "*:GetHighestBlock",
    "/GetMaxHeight": "*:GetMaxHeight",
    "/GetBlocksFrom": "*:GetBlocksFrom",
    "/GetBalance": "*:GetBalance",
    "/GetNonce": "*:GetNonce",
    "/SendTransaction": "*:SendTransaction",
    //"/CreateAccount": "*:CreateAccount",
    //"/SwitchAccount": "*:SwitchAccount",
    //"/GetAccounts": "*:GetAccounts",
    //"/CurrentAccount": "*:CurrentAccount",
    "/GetTransactionsFromBlock": "*:GetTransactionsFromBlock",
    "/SendTransactionsToMainChain": "*:SendTransactionsToMainChain",
}

type Request struct {
    Method string `json:"method"`
    Params string `json:"params"`
}

type Response struct {
    Success bool `json:"success"`
    ErrorMsg string `json:"errMsg"`
    Data interface{} `json:"body"`
}

type MainController struct {
    beego.Controller
    actionName *string
}

func (controller *MainController) Get() {
    controller.Ctx.WriteString("Hello World!")
}

func (controller *MainController) GetAllBlocks() {
    resp := &Response{Success:true}
    controller.Data["json"] = resp

    blocks := database.GetAllBlocks()
    fmt.Println(blocks)
    var body []*BlockWeb
    for _, block := range(blocks) {
        item := ParseBlock(block)
        body = append(body, item)
    }
    resp.Data = body
    fmt.Println(resp)
    controller.ServeJSON()
}

func (controller *MainController) GetBlock() {
    //var height int64
    value := controller.Input().Get("height")
    height, _ := strconv.ParseInt(value, 10, 64)
    //fmt.Println(height)
    resp := &Response{Success:true}
    controller.Data["json"] = resp

    b := database.GetBlock(height)
    block := ParseBlock(b)
    resp.Data = block
    controller.ServeJSON()
}

func (controller *MainController) GetHighestBlock() {
    resp := &Response{Success:true}
    controller.Data["json"] = resp
    block := database.GetHighestBlock()
    resp.Data = ParseBlock(block)
    controller.ServeJSON()
}

func (controller *MainController) GetMaxHeight() {
    resp := &Response{Success:true}
    controller.Data["json"] = resp
    height := database.GetMaxHeight()

    if height == -1 {
        errMsg := "error occurred during database.GetMaxHeight()"
        fmt.Println(errMsg)
        resp.Success = false
        resp.ErrorMsg = errMsg
        resp.Data = height
        controller.ServeJSON()
        return
    }
    resp.Data = height
    controller.ServeJSON()
}

func (controller *MainController) GetBlocksFrom(){
    resp := &Response{Success:true}
    controller.Data["json"] = resp
    st := controller.Input().Get("start")
    sz := controller.Input().Get("size")
    start, err := strconv.ParseInt(st, 10, 64)
    if err != nil {
        resp.ErrorMsg = "error occurred: param start is not an integer."
        controller.ServeJSON()
        return
    }
    size, err := strconv.ParseInt(sz, 10, 64)
    if err != nil {
        resp.Success = false
        resp.ErrorMsg = "error occurred: param size is not an integer."
        controller.ServeJSON()
        return
    }
    blocks := database.GetBlocksFrom(start, size)
    var body []*BlockWeb
    for _, block := range(blocks) {
        item := ParseBlock(block)
        body = append(body, item)
    }
    resp.Data = body
    fmt.Println(resp)
    controller.ServeJSON()
}

func (controller *MainController) GetBalance() {
   // find param in http.Request
    resp := &Response{Success:false}
    controller.Data["json"] = resp
    address := controller.GetString("address")
    if len(address) < 2 {
        resp.ErrorMsg = "param format incorrect"
        return
    }
    address = address[2:]
    fmt.Println("BalanceAddress: ", address)
    ca := accounts.Hex2Address(address)
    c := controller.GetString("chainId")
    chainId := config.Hex2ChainId(c)
    b := database.GetBalance(ca, chainId)
    resp.Success = true
    resp.Data = b.String()

    fmt.Println(resp)
    controller.ServeJSON()
}

func (controller *MainController) GetNonce() {
    resp := &Response{Success:true}
    controller.Data["json"] = resp
    address := controller.Input().Get("address")
    fmt.Println("NonceAddress: ", address)
    ca := accounts.Hex2Address(address)
    c := controller.GetString("chainId")
    chainId := config.Hex2ChainId(c)
    nonce := database.GetNonce(ca, chainId)
    resp.Data = nonce
    controller.ServeJSON()
}

func (controller *MainController) SendTransaction() {
    resp := &Response{Success:false}
    controller.Data["json"] = resp
    to := controller.Input().Get("to")
    a := controller.Input().Get("amount")
    d := controller.Input().Get("destChain")

    amount, succeed := new(big.Int).SetString(a, 10)
    if succeed == false {
        resp.ErrorMsg = "params amount parsing error"
        controller.ServeJSON()
        return
    }

    destChain := config.Hex2ChainId(d)
    t := node.GenerateBalanceTransaction(to, destChain, amount)

    var body string
    if node.SendTransaction(t) != nil {
        body = "offline!"
    } else {
        body = "Send finished!"
    }
    resp.Success = true
    resp.Data = body
    controller.ServeJSON()
}

func (controller *MainController) GetTransactionsFromBlock() {
    resp := &Response{Success:true}
    controller.Data["json"] = resp
    value := controller.Input().Get("height")

    height, err := strconv.ParseInt(value, 10, 64)
    if err != nil {
        resp.Success = false
        resp.ErrorMsg = err.Error()
        controller.ServeJSON()
        return
    }
    block := database.GetBlock(height)
    txs := block.Data.TxList
    var body []*TransactionWeb
    for _, tx := range(txs) {
        tx := ParseTransaction(tx)
        body = append(body, tx)
    }
    resp.Data = body
    controller.ServeJSON()
}


func (controller *MainController) SendTransactionsToMainChain() {
    resp := &Response{Success:true}
    controller.Data["json"] = resp
    value := controller.Input().Get("tx_pkg")
    txsBytes := []byte(value)
    var txs []*bean.Transaction
    err := json.Unmarshal(txsBytes, txs)
    if err != nil {
        resp.Success = false
        resp.ErrorMsg = err.Error()
        controller.ServeJSON()
        return
    }
    resp.Data = "send transactions succeed!"
    controller.ServeJSON()
    go func() {
        for _, tx := range(txs) {
            node.SendTransaction(tx)
        }
    }()
}

func Start() {
    go func() {
        port := flag.String("port", "55550", "port:default is 55551")
        fmt.Println("http server is ready for listen port:", port)
        beego.Router("/", &MainController{})
        for pattern, mappingMethods := range (mappingMethodMap) {
            //fmt.Println(pattern, ": ", mappingMethods)
            beego.Router(pattern, &MainController{}, mappingMethods)
        }

        //beego.Router("/api/list", &MainController{}, "*:List")
        beego.Run(":" + *port)
    }()
}