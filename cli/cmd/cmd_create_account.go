package cmd

import (
    "github.com/spf13/cobra"
    "fmt"
    "BlockChainTest/config"
    "strconv"
)

var create = "create"

var cmdCreate = &cobra.Command{

    Use: create,

    Short: `"` + create + `" is command to create new account.`,

    Long: `"` + create + `" is command to create new account. if --chainid is not specified, chainId will be set to default
value 0, which refers to drep root chain. if you are currently on root chain and want to create root chain account, both --chainid
and --keystore shouldn't be specified. if you are currently on root chain and want to create child chain account, you should specify
--chainid and --keystore, and a new keystore file of a new account of that child chain will be saved under the file path which --keystore
refers to. if you are currently on child chain and want to create child chain account, you should specify --keystore to indicate your
root chain keystore file path, and make sure --chainid is set to the exact chainId value of current chain. if you are currently on child
chain and want to create root chain account, your request will be denied for the operation is not permitted.`,

    Run: func(cmd *cobra.Command, args []string) {
        chainId, _ := strconv.ParseInt(cmd.Flag(flagChainId).Value.String(), 10, 64)
        keystore := cmd.Flag(flagKeystore).Value.String()
        url := urlCreateAccount(chainId, keystore)
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
    cmdCreate.Flags().Int64VarP(&ptrChainId, flagChainId, "c", config.GetChainId(), "chain id")
    cmdCreate.Flags().StringVarP(&ptrKeystore, flagKeystore, "k", "", "keystore")
    CmdRoot.AddCommand(cmdCreate)
}