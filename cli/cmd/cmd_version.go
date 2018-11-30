package cmd

import (
    "github.com/spf13/cobra"
    "fmt"
)

var version = "version"

var cmdVersion = &cobra.Command{

    Use: version,

    Short: `"` + version + `" is command to check current ` + Root + " version.",

    Long: `"` + version + `" is command to check current ` + Root + " version.",

    Run: func(cmd *cobra.Command, args []string) {
        fmt.Println(Root + " " + version + " 1.0")
    },
}

func init() {
    CmdRoot.AddCommand(cmdVersion)
}