package types

type Config struct {
	EnableWallet bool		`json:"enableWallet"`
	KeyStoreDir string		`json:"keyStoreDir,omitempty"`
	WalletPassword string	`json:"walletPassword,omitempty"`
}
