package cmd

import (
    "github.com/spf13/cobra"
    "fmt"
    "encoding/json"
)

var _switch = "switch"
var addr string

var cmdSwitch = &cobra.Command{

    Use: _switch,

    Short: `"` + _switch + `" is the command to switch to another accounts.`,

    Long: `"` + _switch + `" is the command to switch to another accounts.`,

    Run: func(cmd *cobra.Command, args []string) {
        addr = cmd.Flag(flagAccount).Value.String()
        url := urlSwitchAccount(addr)
        data, err := GetRequest(url)
        if err != nil {
            errSwitch(err)
            return
        }

        resp := &Response{}
        err = json.Unmarshal(data, resp)
        if err != nil {
            errSwitch(err)
            return
        }
        if !resp.OK() {
            errSwitch(resp.ErrorMsg)
            return
        }

        fmt.Println("succeed switching to accounts: " + addr)
    },
}

func init() {
    cmdSwitch.Flags().StringVarP(&ptrAccount, flagAccount, "a", "", "switch accounts")
    CmdRoot.AddCommand(cmdSwitch)
}

func errSwitch(err interface{}) {
    fmt.Println("failed to switch to accounts " + addr + ", error: ", err)
}