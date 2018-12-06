package cmd

import (
    "github.com/spf13/cobra"
    "fmt"
)

var maxHeight = "maxheight"

var cmdMaxHeight = &cobra.Command{

    Use: maxHeight,

    Short: `"` + maxHeight + `" is command to check current maximum block height`,

    Long: `"` + maxHeight + `" is command to check current maximum block height`,

    Run: func(cmd *cobra.Command, args []string) {
        url := urlMaxHeight()
        resp, err := GetResponse(url)
        if err != nil {
            fmt.Println(ErrCheckMaxHeight, err)
            return
        }
        if !resp.Success {
            fmt.Println(ErrCheckMaxHeight, err)
            return
        }
        fmt.Println("max height: ", resp.Body)
    },
}

func init() {
    CmdRoot.AddCommand(cmdMaxHeight)
}
