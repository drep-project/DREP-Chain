package main

import (
	"fmt"
	"gopkg.in/urfave/cli.v1"
	"BlockChainTest/util/flags"
	"BlockChainTest/config"
	"BlockChainTest/accounts"
)
var (
	accountCommand = cli.Command{
		Name:     "account",
		Usage:    "Manage accounts",
		Category: "ACCOUNT COMMANDS",
		Description: `
Manage accounts, list all existing accounts, import a private key into a new
account, create a new account or update an existing account.

It supports interactive mode, when you are prompted for password as well as
non-interactive mode where passwords are supplied via a given password file.
Non-interactive mode is only meant for scripted use on test networks or known
safe environments.

Make sure you remember the password you gave when creating a new account (with
either new or import). Without it you are not able to unlock your account.

Note that exporting your key in unencrypted format is NOT supported.

Keys are stored under <DATADIR>/keystore.
It is safe to transfer the entire directory or the individual keys therein
between drep nodes by simply copying.

Make sure you backup your keys regularly.`,
		Subcommands: []cli.Command{
			{
				Name:   "list",
				Usage:  "Print summary of existing accounts",
				Action: flags.MigrateFlags(accountList),
				Flags: []cli.Flag{
					flags.DataDirFlag,
					flags.KeyStoreDirFlag,
				},
				Description: `
Print a short summary of all accounts`,
			},
			{
				Name:   "new",
				Usage:  "Create a new account",
				Action: flags.MigrateFlags(accountCreate),
				Flags: []cli.Flag{
					flags.DataDirFlag,
					flags.KeyStoreDirFlag,
				},
				Description: `
    geth account new

Creates a new account and prints the address.

The account is saved in encrypted format, you are prompted for a passphrase.

You must remember this passphrase to unlock your account in the future.

For non-interactive use the passphrase can be specified with the --password flag:

Note, this is meant to be used for testing only, it is a bad idea to save your
password to file or expose in any other way.
`,
			},
		},
	}
)

func accountList(ctx *cli.Context) error {
	nCfg, err := config.MakeConfig(ctx)
	if err != nil {
		return err
	}
	api := accounts.AccountApi{
		KeyStoreDir : nCfg.Keystore,
		ChainId : config.Hex2ChainId(nCfg.ChainId),
	}
	address, err := api.AccountList()
	if err != nil {
		return err
	}
	fmt.Println(address)
	return nil
}

// accountCreate creates a new account into the keystore defined by the CLI flags.
func accountCreate(ctx *cli.Context) error {
	nCfg, err := config.MakeConfig(ctx)
	if err != nil {
		return err
	}
	
	api := accounts.AccountApi{
		KeyStoreDir : nCfg.Keystore,
		ChainId : config.Hex2ChainId(nCfg.ChainId),
	}
	newAddress, err := api.CreateAccount()
	if err != nil {
		return err
	}
	fmt.Println(newAddress)
	return nil
}