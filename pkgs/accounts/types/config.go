package types

type Config struct {
	Enable      bool   `json:"enable"`
	Type        string `json:"type,omitempty"`
	KeyStoreDir string `json:"keyStoreDir,omitempty"`
	Password    string `json:"password,omitempty"`
}
