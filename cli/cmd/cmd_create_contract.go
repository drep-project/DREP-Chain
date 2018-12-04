package cmd

import (
    "github.com/spf13/cobra"
    "fmt"
)

var createContract = "create-contract"

var cmdCreateContract = &cobra.Command{

    Use: createContract,

    Short: `"` + createContract + `" is command to create new smart contract.`,

    Long: `"` + createContract + `" is command to create new smart contract. --code should be set as a hex string or omitted.
if --code is set, a smart contract of that code will be created.`,

    Run: func(cmd *cobra.Command, args []string) {
        code := cmd.Flag(flagCode).Value.String()
        url := urlSendCreateContractTransaction(code)
        resp, err := GetResponse(url)
        if err != nil {
            fmt.Println(ErrCreateContract, err)
            return
        }
        if !resp.Success {
            fmt.Println(ErrCreateContract, resp.ErrorMsg)
            return
        }
        fmt.Println("succeed creating new smart contract ", resp.Body)
    },
}

func init() {
    cmdCreateContract.Flags().StringVarP(&ptrCode, flagCode, "c", "", "byte code")
    CmdRoot.AddCommand(cmdCreateContract)
}
