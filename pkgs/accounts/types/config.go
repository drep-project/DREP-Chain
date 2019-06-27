package types

type Config struct {
	Enable      bool   `json:"enable"`
	Type        string `json:"enable"`
	KeyStoreDir string `json:"keyStoreDir,omitempty"`
	Password    string `json:"password,omitempty"`
}
