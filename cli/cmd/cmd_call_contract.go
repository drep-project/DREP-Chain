package cmd

import (
    "github.com/spf13/cobra"
    "fmt"
)

var callContract = "call-contract"

var cmdCallContract = &cobra.Command{

    Use: callContract,

    Short: `"` + createContract + `" is command to call smart contract.`,

    Long: `"` + createContract + `" is command to call smart contract.`,

    Run: func(cmd *cobra.Command, args []string) {
        addr := cmd.Flag(flagAccount).Value.String()
        chainId := cmd.Flag(flagChainId).Value.String()
        input := cmd.Flag(flagInput).Value.String()
        readOnly := cmd.Flag(flagReadOnly).Value.String()
        url := urlSendCallContractTransaction(addr, chainId, input, readOnly)
        resp, err := GetResponse(url)
        if err != nil {
            fmt.Println(ErrCallContract, err)
            return
        }
        if !resp.Success {
            fmt.Println(ErrCallContract, resp.ErrorMsg)
            return
        }
        fmt.Println("succeed calling smart contract")
    },
}

func init() {
    cmdCallContract.Flags().StringVarP(&ptrAccount, flagAccount, "a", "", "contract address")
    cmdCallContract.Flags().Int64VarP(&ptrChainId, flagChainId, "c", 0, "chain id")
    cmdCallContract.Flags().StringVarP(&ptrInput, flagInput, "i", "", "contract input")
    cmdCallContract.Flags().BoolVarP(&ptrReadOnly, flagReadOnly, "r", false, "read only mark")
    CmdRoot.AddCommand(cmdCallContract)
}
