package cmd

import (
    "github.com/spf13/cobra"
    "fmt"
    "encoding/json"
    "strconv"
)

var block = "block"

var cmdBlock = &cobra.Command{

    Use: block,

    Short: `"` + block + `" is the command to fetch and print local block details`,

    Long: `"` + block + `" is the command to fetch and print local block details, if --height, --begin, --size are all omitted, 
all blocks will be returned; if --height is specified, only the block of that specific height will be returned; otherwise, if both 
--begin and --size are specified, the blocks of height from "begin" to "begin + size - 1" will be returned; if only --size are 
specified, only the most recent "size" number of blocks will be returned; if only --begin are specified, all the most recent blocks of 
height starting from the "begin" will be returned.`,

    Run: func(cmd *cobra.Command, args []string) {
        var url string
        height, _ := strconv.ParseInt(cmd.Flag(flagHeight).Value.String(), 10, 64)
        begin, _ := strconv.ParseInt(cmd.Flag(flagBegin).Value.String(), 10, 64)
        size, _ := strconv.ParseInt(cmd.Flag(flagSize).Value.String(), 10, 64)
        if height > -1 {
            url = urlBlock(height)
        } else if begin == -1 && size == -1 {
            url = urlAllBlocks()
        } else if begin == -1 {
            url = urlMostRecentBlocks(size)
        } else {
            url = urlBlocksFrom(begin, size)
        }

        data, err := GetRequest(url)
        if err != nil {
            errBlock(err)
            return
        }

        resp := &Response{}
        err = json.Unmarshal(data, resp)
        if err != nil {
            errBlock(err)
            return
        }
        if !resp.OK() {
            errBlock(resp.ErrorMsg)
            return
        }

        fmt.Println(resp.Body)
    },
}

func init() {
    cmdBlock.Flags().Int64VarP(&ptrHeight, flagHeight, "H", -1, "height of the fetched block")
    cmdBlock.Flags().Int64VarP(&ptrBegin, flagBegin, "b", -1, "starting height of fetched blocks")
    cmdBlock.Flags().Int64VarP(&ptrSize, flagSize, "s", -1, "number of fetched blocks")
    CmdRoot.AddCommand(cmdBlock)
}

func errBlock(err interface{}) {
    fmt.Println("get blocks error: ", err)
}