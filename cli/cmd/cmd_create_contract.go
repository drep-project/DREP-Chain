package cmd

import (
    "github.com/spf13/cobra"
    "fmt"
)

var createContract = "create-contract"

var cmdCreateContract = &cobra.Command{

    Use: createContract,

    Short: `"` + createContract + `" is command to create new smart contract.`,

    Long: `"` + createContract + `" is command to create new smart contract.`,

    Run: func(cmd *cobra.Command, args []string) {
        codeFile := cmd.Flag(flagCodeFile).Value.String()
        url := urlSendCreateContractTransaction(codeFile)
        resp, err := GetResponse(url)
        if err != nil {
            fmt.Println(ErrCreateAccount, err)
            return
        }
        if !resp.Success {
            fmt.Println(ErrCreateAccount, resp.ErrorMsg)
            return
        }
        fmt.Println("succeed creating new accounts: ", resp.Body)
    },
}

func init() {
    cmdCreateContract.Flags().StringVarP(&ptrCodeFile, flagCodeFile, "f", "", "byte code storage file")
    CmdRoot.AddCommand(cmdCreate)
}
