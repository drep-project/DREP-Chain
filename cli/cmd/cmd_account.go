package cmd

import (
    "github.com/spf13/cobra"
    "fmt"
)

var accounts = "account"

var cmdAccounts = &cobra.Command{

    Use: accounts,

    Short: `"` + accounts + `" is command to check account address.`,

    Long: `"` + accounts + `" is command to check account address on current chain.`,

    Run: func(cmd *cobra.Command, args []string) {
        url := urlGetAccount()
        resp, err := GetResponse(url)
        if err != nil {
            fmt.Println(ErrGetAccount, err)
        }
        if !resp.Success {
            fmt.Println(ErrGetAccount, resp.ErrorMsg)
            return
        }
        if resp.Body == "" {
            fmt.Println(ErrGetAccount, "no account found on current chain")
            return
        }
        fmt.Println(resp.Body)
    },
}

func init() {
    CmdRoot.AddCommand(cmdAccounts)
}