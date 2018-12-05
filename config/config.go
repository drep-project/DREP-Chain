package config

import (
    "github.com/spf13/viper"
    "fmt"
    "os"
)

const (
    RootChain = 0
)

var (
    dataDir = "." + string(os.PathSeparator) + "data"
)

type Config struct {
    RelayNode  []string
    ChainId    int64
    ConfigDir  string
    DataDir    string
    DocsDir    string
    Keystore   string
    IP         string
    Port       int
    MyIndex    int
    DebugNodes []*DebugNode
}

func init() {
    viper.SetConfigName("config")
    viper.AddConfigPath(dataDir)
    err := viper.ReadInConfig()
    if err != nil {
        fmt.Println("read config file error: ", err)
    }
}

func GetChainId() int64 {
    return viper.GetInt64("ChainId")
}

func GetKeystore() string {
    return dataDir + string(os.PathSeparator) + "keystore"
}

func GetMyIndex() int {
    return viper.GetInt("MyIndex")
}

func GetDebugNodes() []*DebugNode {
    config := &Config{}
    viper.Unmarshal(config)
    return config.DebugNodes
}

func SetChain(chainId int64, dataDir string) error {
    viper.Set("ChainId", chainId)
    viper.Set("DataDir", dataDir)
    return viper.WriteConfig()
}

func SetKeystore(keystorePath string) error {
    viper.Set("Keystore", keystorePath)
    return viper.WriteConfig()
}