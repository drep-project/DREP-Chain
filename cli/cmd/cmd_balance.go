package cmd

import (
    "github.com/spf13/cobra"
    "fmt"
    "encoding/json"
)

var balance = "balance"

var cmdBalance = &cobra.Command{

    Use: balance,

    Short: `"` + balance + `" is the command to check current balance of a specific account.`,

    Long: `"` + balance + `" is the command to check current balance a specific account. if "--` + flagAccount +
        `" is set to a specific account address, the balance of that account will be returned if the address is valid, otherwise 
 error will be returned.`,

    Run: func(cmd *cobra.Command, args []string) {
        addr := cmd.Flag(flagAccount).Value.String()
        url := urlBalance(addr)
        data, err := GetRequest(url)
        if err != nil {
            errBalance(err)
            return
        }

        resp := &Response{}
        err = json.Unmarshal(data, resp)
        if err != nil {
            errBalance(err)
            return
        }
        if !resp.OK() {
            errBalance(resp.ErrorMsg)
            return
        }

        fmt.Println("balance:" , resp.Body)
    },

}

func init() {
    cmdBalance.Flags().StringVarP(&ptrAccount, flagAccount, "a", "", "account address")
    CmdRoot.AddCommand(cmdBalance)
}

func errBalance(err interface{}) {
    fmt.Println("check balance error: ", err)
}

