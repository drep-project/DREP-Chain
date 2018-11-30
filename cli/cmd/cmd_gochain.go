package cmd

import (
    "github.com/spf13/cobra"
    "fmt"
    "BlockChainTest/config"
)

var goChain = "gochain"

var cmdGoChain = &cobra.Command{

    Use: goChain,

    Short: `"` + goChain + `" is command to start running chain.`,

    Long: `"` + create + `" is command to start running chain. if --chainid is not specified, it will be set to default
value 0 which refers to drep root chain. chain with id of --chainid value will start running. all chain data will be stored
into directory which --datadir refers to`,

    Run: func(cmd *cobra.Command, args []string) {
        chainId := cmd.Flag(flagChainId).Value.String()
        dataDir := cmd.Flag(flagDataDir).Value.String()
        url := urlSetChain(chainId, dataDir)
        resp, err := GetResponse(url)
        if err != nil {
            fmt.Println(ErrGoChain, err)
            return
        }
        if err != nil {
            fmt.Println(ErrGoChain, err)
            return
        }
        if !resp.Success {
            fmt.Println(ErrGoChain, resp.ErrorMsg)
            return
        }
        fmt.Println("succeed! currently running chain: ", chainId)
    },
}

func init() {
    cmdGoChain.Flags().Int64VarP(&ptrChainId, flagChainId, "c", config.GetChainId(), "chain id")
    cmdGoChain.Flags().StringVarP(&ptrDataDir, flagDataDir, "d", "", "data dir")
    CmdRoot.AddCommand(cmdCreate)
}