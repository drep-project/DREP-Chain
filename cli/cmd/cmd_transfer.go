package cmd

import (
    "github.com/spf13/cobra"
    "fmt"
)

var transfer = "transfer"

var cmdSend = &cobra.Command{

    Use: transfer,

    Short: `"` + transfer + `" is command to send transfer transaction.`,

    Long: `"` + transfer + `" is command to send transfer transaction.`,

    Run: func(cmd *cobra.Command, args []string) {
        to := cmd.Flag(flagTo).Value.String()
        destChain := cmd.Flag(flagDestChain).Value.String()
        amount := cmd.Flag(flagAmount).Value.String()
        url := urlSendTransferTransaction(to, destChain, amount)
        resp, err := GetResponse(url)
        if err != nil {
            fmt.Println(ErrSendTransaction, err)
            return
        }
        if !resp.Success {
            fmt.Println(ErrSendTransaction, resp.ErrorMsg)
            return
        }
        fmt.Println("succeed sending transfer transaction")
    },
}

func init() {
    cmdSend.Flags().StringVarP(&ptrTo, flagTo, "t", "", "receiver address")
    cmdSend.Flags().Int64VarP(&ptrDestChain, flagDestChain, "d", 0, "receiver chain id")
    cmdSend.Flags().StringVarP(&ptrAmount, flagAmount, "a", "", "transfer amount")
    CmdRoot.AddCommand(cmdSend)
}