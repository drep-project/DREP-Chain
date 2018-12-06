package cmd

import (
    "github.com/spf13/cobra"
    "fmt"
    "BlockChainTest/config"
)

var nonce = "nonce"

var cmdNonce = &cobra.Command{

    Use: nonce,

    Short: `"` + nonce + `is command to check nonce of current account.`,

    Long: `"` + nonce + `is command to check nonce of current account.`,

    Run: func(cmd *cobra.Command, args []string) {
        address := cmd.Flag(flagAccount).Value.String()
        chainId := cmd.Flag(flagChainId).Value.String()
        url := urlNonce(address, chainId)
        resp, err := GetResponse(url)
        if err != nil {
            fmt.Println(ErrCheckNonce, err)
            return
        }
        if !resp.Success {
            fmt.Println(ErrCheckNonce, resp.ErrorMsg)
            return
        }
        fmt.Println("nonce: ", resp.Body)
    },
}

func init() {
    cmdBalance.Flags().StringVarP(&ptrAccount, flagAccount, "a", "", "account address")
    cmdBalance.Flags().Int64VarP(&ptrChainId, flagChainId, "c",  config.GetChainId(), "chain id")
    CmdRoot.AddCommand(cmdNonce)
}