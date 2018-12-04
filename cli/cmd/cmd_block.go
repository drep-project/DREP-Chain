package cmd

import (
    "github.com/spf13/cobra"
    "fmt"
    "strconv"
    "BlockChainTest/bean"
    "encoding/json"
)

var block = "block"

var cmdBlock = &cobra.Command{

    Use: block,

    Short: `"` + block + `" is the command to fetch and print local block details`,

    Long: `"` + block + `" is the command to fetch and print local block details, if --height is specified, only the block 
of that specific height will be returned; otherwise, --height will be set to default value 0 and most recent blocks are returned.`,

    Run: func(cmd *cobra.Command, args []string) {
        height, _ := strconv.ParseInt(cmd.Flag(flagHeight).Value.String(), 10, 64)
        if height < -1 {
            fmt.Println(ErrGetBlock, "--height value should be positive ")
            return
        }
        url := urlGetBlock(height)
        resp, err := GetResponse(url)
        if err != nil {
            fmt.Println(ErrGetBlock, err)
        }
        if !resp.Success {
            fmt.Println(ErrGetBlock, resp.ErrorMsg)
        }
        block := &bean.Block{}
        err = json.Unmarshal([]byte(resp.Body.(string)), block)
        if err != nil {
            fmt.Println(ErrGetBlock, err)
            return
        }
        fmt.Println("block info:")
        fmt.Println("version: ", block.Header.Version)
        fmt.Println("height: ", block.Header.Height)
        fmt.Println("chain id: ", block.Header.ChainId)
        fmt.Println("previous hash: ", block.Header.PreviousHash)
        fmt.Println("gas limit: ", block.Header.GasLimit)
        fmt.Println("gas used: ", block.Header.GasUsed)
        fmt.Println("timestamp: ", block.Header.Timestamp)
        fmt.Println("transaction hash: ", block.Header.TxHashes)
        fmt.Println("merkle root: ", block.Header.MerkleRoot)
    },
}

func init() {
    cmdBlock.Flags().Int64VarP(&ptrHeight, flagHeight, "h", -1, "block height")
    CmdRoot.AddCommand(cmdBlock)
}