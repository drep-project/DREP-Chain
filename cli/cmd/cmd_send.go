package cmd

import (
    "github.com/spf13/cobra"
    "fmt"
    "encoding/json"
)

var send = "send"

var cmdSend = &cobra.Command{

    Use: send,

    Short: `"` + send + `" is the command to send drep coins to another account.`,

    Long: `"` + send + `" is the command to send drep coins to another account.`,

    Run: func(cmd *cobra.Command, args []string) {
        to := cmd.Flag(flagTo).Value.String()
        amount := cmd.Flag(flagAmount).Value.String()
        url := urlSendTransaction(to, amount)
        data, err := GetRequest(url)
        if err != nil {
            errSend(err)
            return
        }

        resp := &Response{}
        err = json.Unmarshal(data, resp)
        if err != nil {
            errSend(err)
            return
        }
        if !resp.OK() {
            errSend(resp.ErrorMsg)
            return
        }

        fmt.Println("succeed sending transaction")
    },
}

func init() {
    cmdSend.Flags().StringVarP(&ptrTo, flagTo, "t", "", "receiver address")
    cmdSend.Flags().StringVarP(&ptrAmount, flagAmount, "a", "", "transfer amount")
    CmdRoot.AddCommand(cmdSend)
}

func errSend(err interface{}) {
    fmt.Println("failed to send transaction, error: ", err)
}