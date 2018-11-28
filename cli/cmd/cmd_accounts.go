package cmd

import (
    "github.com/spf13/cobra"
    "fmt"
    "encoding/json"
    "strconv"
)

var accounts = "accounts"

var cmdAccounts = &cobra.Command{

    Use: accounts,

    Short: `"` + accounts + `" is the command to check all local accounts reserved.`,

    Long: `"` + accounts + `" is the command to check and return all local accounts reserved.`,

    Run: func(cmd *cobra.Command, args []string) {
        url := urlAccounts()
        data, err := GetRequest(url)
        if err != nil {
            errAccounts(err)
            return
        }

        resp := &Response{}
        err = json.Unmarshal(data, resp)
        if err != nil {
            errAccounts(err)
            return
        }
        if !resp.OK() {
            errAccounts(resp.ErrorMsg)
            return
        }

        if ret, ok := resp.Body.([]interface{}); ok {
            if ret == nil {
                errAccounts("invalid server response")
                return
            }
            hint := "find " + strconv.FormatInt(int64(len(ret)), 10) + " accounts"
            if len(ret) < 2 {
                hint = hint[: len(hint) - 1]
            }
            fmt.Println(hint)
            for i, b := range ret {
              fmt.Println(i + 1, ": ", b)
            }
        } else {
            errAccounts("invalid server response")
        }
    },
}

func init() {
    CmdRoot.AddCommand(cmdAccounts)
}

func errAccounts(err interface{}) {
    fmt.Println("failed to check accounts, error: ", err)
}
