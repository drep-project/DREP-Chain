package cmd

import (
    "github.com/spf13/cobra"
    "fmt"
    "encoding/json"
)

var create = "create"

var cmdCreate = &cobra.Command{

    Use: create,

    Short: `"` + create + `" is the command to create new accounts.`,

    Long: `"` + create + `" is the command to create new accounts.`,

    Run: func(cmd *cobra.Command, args []string) {
        url := urlCreateAccount()
        data, err := GetRequest(url)
        if err != nil {
            errCreate(err)
            return
        }

        resp := &Response{}
        err = json.Unmarshal(data, resp)
        if err != nil {
            errCreate(err)
            return
        }
        if !resp.OK() {
            errCreate(resp.ErrorMsg)
            return
        }

        fmt.Println("succeed creating new accounts: ", resp.Body)
    },
}

func init() {
    CmdRoot.AddCommand(cmdCreate)
}

func errCreate(err interface{}) {
    fmt.Println("failed to create new accounts，error: ", err)
}