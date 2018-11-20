package cmd

import (
    "github.com/spf13/cobra"
    "fmt"
    "encoding/json"
)

var current = "current"

var cmdCurrent = &cobra.Command{

    Use: current,

    Short: `"` + current + `" is the command to check current account address.`,

    Long: `"` + current + `" is the command to check and return current account address.`,

    Run: func(cmd *cobra.Command, args []string) {
        url := urlCurrentAccount()
        data, err := GetRequest(url)
        if err != nil {
            errCurrent(err)
            return
        }

        resp := &Response{}
        err = json.Unmarshal(data, resp)
        if err != nil {
            errCurrent(err)
            return
        }
        if !resp.OK() {
            errCurrent(resp.ErrorMsg)
            return
        }

        fmt.Println(resp.Body)
    },
}

func init() {
    CmdRoot.AddCommand(cmdCurrent)
}

func errCurrent(err interface{}) {
    fmt.Println("failed to check current account address, error: ", err)
}