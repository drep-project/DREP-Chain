package cmd

import (
    "github.com/spf13/cobra"
    "fmt"
    "BlockChainTest/config"
)

var balance = "balance"

var cmdBalance = &cobra.Command{

    Use: balance,

    Short: `"` + balance + `" is command to check balance of current account.`,

    Long: `"` + balance + `" is command to check balance of current account.`,

    Run: func(cmd *cobra.Command, args []string) {
        address := cmd.Flag(flagAccount).Value.String()
        chainId := cmd.Flag(flagChainId).Value.String()
        url := urlBalance(address, chainId)
        resp, err := GetResponse(url)
        if err != nil {
            fmt.Println(ErrCheckBalance, err)
            return
        }
        if !resp.Success {
            fmt.Println(ErrCheckBalance, resp.ErrorMsg)
            return
        }
        fmt.Println("balance:" , resp.Body)
    },

}

func init() {
    cmdBalance.Flags().StringVarP(&ptrAccount, flagAccount, "a", "", "account address")
    cmdBalance.Flags().Int64VarP(&ptrChainId, flagChainId, "c",  config.GetChainId(), "chain id")
    CmdRoot.AddCommand(cmdBalance)
}