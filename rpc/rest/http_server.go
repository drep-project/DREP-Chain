package rest

import (
    "fmt"
    "flag"
    "strconv"
    "encoding/json"
    "math/big"
    "github.com/spf13/viper"
    "github.com/astaxie/beego"
    "BlockChainTest/node"
    "BlockChainTest/database"
    "BlockChainTest/accounts"
<<<<<<< HEAD
    "BlockChainTest/bean"
=======
    "BlockChainTest/config"
>>>>>>> 39bb07a... modify chainId type and add revert
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
    "/GetReputation": "*:GetReputation",
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

<<<<<<< HEAD
type MainController struct {
    beego.Controller
    actionName *string
}
=======
func GetAllBlocks(w http.ResponseWriter, _ *http.Request) {
    fmt.Println("get all blocks running")
    blocks := database.GetAllBlocks()
>>>>>>> 39bb07a... modify chainId type and add revert

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

<<<<<<< HEAD
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

func (controller *MainController)GetHighestBlock() {
    resp := &Response{Success:true}
    controller.Data["json"] = resp
    block := database.GetHighestBlock()
    resp.Data = ParseBlock(block)
    controller.ServeJSON()
}

func (controller *MainController)GetMaxHeight() {
    resp := &Response{Success:true}
    controller.Data["json"] = resp
    height := database.GetMaxHeight()

=======
func GetBlock(w http.ResponseWriter, r *http.Request) {
    params := analysisReqParam(r)
    var height int64
    if value, ok := params["height"].(string); ok {
        height, _ = strconv.ParseInt(value, 10, 64)
    }

    block := database.GetBlock(height)
    if block == nil {
        errMsg := "block is nil"
        fmt.Println(errMsg)
        resp := &Response{Success:false, Data:errMsg}
        writeResponse(w, resp)
        return
    }
    blockWeb := ParseBlock(block)
    resp := &Response{Success:true, Data:blockWeb}
    writeResponse(w, resp)
}

func GetHighestBlock(w http.ResponseWriter, _ *http.Request) {
    block := database.GetHighestBlock()
    if block == nil {
        errMsg := "error occurred during database.GetHighestBlock"
        fmt.Println(errMsg)
        resp := &Response{Success:false, Data:errMsg}
        writeResponse(w, resp)
        return
    }
    blockWeb := ParseBlock(block)
    resp := &Response{Success:true, Data:blockWeb}
    writeResponse(w, resp)
}

func GetMaxHeight(w http.ResponseWriter, _ *http.Request) {
    height := database.GetMaxHeight()
>>>>>>> 39bb07a... modify chainId type and add revert
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

func (controller *MainController)GetBlocksFrom(){
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
<<<<<<< HEAD
=======

>>>>>>> 39bb07a... modify chainId type and add revert
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

func (controller *MainController)GetBalance() {
    // find param in http.Request
<<<<<<< HEAD
    resp := &Response{Success:false}
    controller.Data["json"] = resp
    address := controller.GetString("address")
    address = address[2:]
=======
    params := analysisReqParam(r)
    var address string
    var chainId config.ChainIdType
    if value, ok := params["address"].(string); ok {
        address = value[2:]
    }
    if value, ok := params["chainId"].(string); ok {
        chainId = config.Hex2ChainId(value)
    }
>>>>>>> 39bb07a... modify chainId type and add revert

    if len(address) == 0 {
        resp.ErrorMsg = "param format incorrect"
        return
    }
    c := controller.GetString("chainId")
    chainId, err := strconv.ParseInt(c, 10, 64)
    if err != nil {
        resp.ErrorMsg = err.Error()
        resp.Data = c
        controller.ServeJSON()
        return
    }

    fmt.Println("BalanceAddress: ", address)
    ca := accounts.Hex2Address(address)
<<<<<<< HEAD
    b := database.GetBalanceOutsideTransaction(ca, chainId)
    resp.Success = true
    resp.Data = b.String()

    controller.ServeJSON()
}

func (controller *MainController)GetNonce() {
    resp := &Response{Success:true}
    controller.Data["json"] = resp
    address := controller.Input().Get("address")
    address = address[2:]
    fmt.Println("NonceAddress: ", address)
=======
    //database.PutBalance(ca, big.NewInt(1314))
    //fmt.Println("[database PutBalance] succeed!")

    b := database.GetBalance(ca, chainId)
    resp := &Response{Success:true, Data:b.String()}
    writeResponse(w, resp)
}

func GetNonce(w http.ResponseWriter, r *http.Request) {
    params := analysisReqParam(r)
    var address string
    var chainId config.ChainIdType
    if value, ok := params["address"].(string); ok {
        address = value
    }
    if value, ok := params["chainId"].(string); ok {
        chainId = config.Hex2ChainId(value)
    }
>>>>>>> 39bb07a... modify chainId type and add revert

    c := controller.Input().Get("chainId")
    chainId, err := strconv.ParseInt(c, 10, 64)
    if err != nil {
        resp.Success = false
        resp.ErrorMsg = err.Error()
        controller.ServeJSON()
        return
    }

    ca := accounts.Hex2Address(address)
<<<<<<< HEAD
    nonce := database.GetNonceOutsideTransaction(ca, chainId)
    resp.Data = nonce
    controller.ServeJSON()
}

func (controller *MainController)SendTransaction() {
    resp := &Response{Success:false}
    controller.Data["json"] = resp
    to := controller.Input().Get("to")
    to = to[2:]
    a := controller.Input().Get("amount")
    d := controller.Input().Get("destChain")
=======

    nonce := database.GetNonce(ca, chainId)
    resp := &Response{Success:true, Data:nonce}
    writeResponse(w, resp)
}

func SendTransaction(w http.ResponseWriter, r *http.Request) {
    params := analysisReqParam(r)
    var to string
    var amount string
    var destChain config.ChainIdType
    if value, ok := params["to"].(string); ok {
        to = value[2:]
    }
    if value, ok := params["amount"].(string); ok {
        amount = value
    }
    if value, ok := params["destChain"].(string); ok {
        destChain = config.Hex2ChainId(value)
    }
>>>>>>> 39bb07a... modify chainId type and add revert

    amount, succeed := new(big.Int).SetString(a, 10)
    if succeed == false {
        resp.ErrorMsg = "params amount parsing error"
        controller.ServeJSON()
        return
    }

    destChain, err := strconv.ParseInt(d, 10, 64)
    if err != nil {
        resp.ErrorMsg = err.Error()
        controller.ServeJSON()
        return
    }

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

func (controller *MainController)GetTransactionsFromBlock() {
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


func (controller *MainController)SendTransactionsToMainChain() {
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

func (controller *MainController)GetReputation() {
    // find param in http.Request
    resp := &Response{Success:false}
    controller.Data["json"] = resp
    address := controller.GetString("address")
    address = address[2:]

    if len(address) == 0 {
        resp.ErrorMsg = "param format incorrect"
        return
    }

    c := controller.GetString("chainId")
    chainId, err := strconv.ParseInt(c, 10, 64)
    if err != nil {
        resp.ErrorMsg = err.Error()
        resp.Data = c
        controller.ServeJSON()
        return
    }

    fmt.Println("BalanceAddress: ", address)
    ca := accounts.Hex2Address(address)
    b := database.GetReputationOutsideTransaction(ca, chainId)
    if (b.Int64() == 0) {
        defaultRep := viper.GetInt64("default_rep")
        fmt.Println("default reputation is :", defaultRep)
        database.PutReputationOutSideTransaction(ca, chainId, big.NewInt(defaultRep))
        resp.Success = true
        resp.Data = viper.GetString("default_rep")
        controller.ServeJSON()
    }
    resp.Success = true
    resp.Data = b.String()

    fmt.Println(resp)
    controller.ServeJSON()
}

func Start() *MainController{
    controller := &MainController{}
    port := flag.String("port", "55550", "port:default is 55551")
    fmt.Println("http server is ready for listen port:", port)
    beego.Router("/", controller)
    for pattern, mappingMethods := range (mappingMethodMap) {
        //fmt.Println(pattern, ": ", mappingMethods)
        beego.Router(pattern, controller, mappingMethods)
    }
    //beego.Router("/api/list", &MainController{}, "*:List")
    beego.Run(":" + *port)
    return controller
}
