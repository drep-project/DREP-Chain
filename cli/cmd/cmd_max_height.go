package cmd

import (
    "github.com/spf13/cobra"
    "fmt"
    "encoding/json"
)

var maxHeight = "maxHeight"

var cmdMaxHeight = &cobra.Command{

    Use: maxHeight,

    Short: `"` + maxHeight + `" is the command to check current maximum block height`,

    Long: `"` + maxHeight + `" is the command to check current maximum block height and will return the height of type int64`,

    Run: func(cmd *cobra.Command, args []string) {
        url := urlMaxHeight()
        data, err := GetRequest(url)
        if err != nil {
            errMaxHeight(err)
            return
        }

        resp := &Response{}
        err = json.Unmarshal(data, resp)
        if err != nil {
            errMaxHeight(err)
            return
        }
        if !resp.OK() {
            errMaxHeight(resp.ErrorMsg)
            return
        }

        fmt.Println("max height: ", resp.Body)
    },
}

func init() {
    CmdRoot.AddCommand(cmdMaxHeight)
}

func errMaxHeight(err interface{}) {
    fmt.Println("check max height error: ", err)
}