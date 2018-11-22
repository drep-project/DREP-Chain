package cmd

import (
    "github.com/spf13/cobra"
    "fmt"
    "encoding/json"
)

var nonce = "nonce"

var cmdNonce = &cobra.Command{

    Use: nonce,

    Short: `"` + nonce + `" is the command to check current nonce(s) of local accounts reserved.`,

    Long: `"` + nonce + `" is the command to check current nonce(s) of local accounts reserved. if "--` + flagAccount +
        `" is set to a specific accounts address, the nonce of that accounts will be returned if the address is valid, otherwise 
error will be returned.`,

    Run: func(cmd *cobra.Command, args []string) {
        addr := cmd.Flag(flagAccount).Value.String()
        url := urlNonce(addr)
        data, err := GetRequest(url)
        if err != nil {
            errNonce(err)
            return
        }

        resp := &Response{}
        err = json.Unmarshal(data, resp)
        if err != nil {
            errNonce(err)
            return
        }
        if !resp.OK() {
            errNonce(resp.ErrorMsg)
            return
        }

        fmt.Println("nonce: ", resp.Body)
    },
}

func init() {
    cmdNonce.Flags().StringVarP(&ptrAccount, flagAccount, "a", "", "accounts address")
    CmdRoot.AddCommand(cmdNonce)
}

func errNonce(err interface{}) {
    fmt.Println("check nonce error: ", err)
}