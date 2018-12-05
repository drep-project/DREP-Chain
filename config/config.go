package config

import (
    "github.com/spf13/viper"
    "fmt"
)

const (
    RootChain = 0
)

var (
    vip *viper.Viper
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
    MinerNum   int
    MyIndex    int
    DebugNodes []*DebugNode
}

func init() {
    if vip == nil {
        vip = viper.New()
        vip.SetConfigName("config")
        vip.AddConfigPath("./config")
    }
    err := vip.ReadInConfig()
    if err != nil {
        fmt.Println("read config file error: ", err)
    }
}

func GetChainId() int64 {
    return vip.GetInt64("ChainId")
}

func GetConfigDir() string {
    return vip.GetString("ConfigDir")
}

func GetDataDir() string {
    return vip.GetString("DataDir")
}

func GetDocsDir() string {
    return vip.GetString("DocsDir")
}

func GetKeystore() string {
    return vip.GetString("Keystore")
}

func GetMinerNum() int {
    return vip.GetInt("MinerNum")
}

func GetMyIndex() int {
    return vip.GetInt("MyIndex")
}

func GetDebugNodes() []*DebugNode {
    config := &Config{}
    vip.Unmarshal(config)
    return config.DebugNodes
}

func SetChain(chainId int64, dataDir string) error {
    vip.Set("ChainId", chainId)
    vip.Set("DataDir", dataDir)
    return vip.WriteConfig()
}

func SetKeystore(keystorePath string) error {
    vip.Set("Keystore", keystorePath)
    return vip.WriteConfig()
}