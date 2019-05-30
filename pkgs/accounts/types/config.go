package types

type Config struct {
	Enable bool				`json:"enable"`
	KeyStoreDir string		`json:"keyStoreDir,omitempty"`
	Password string			`json:"password,omitempty"`
}
