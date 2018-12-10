package config

import (
    "github.com/spf13/viper"
    "fmt"
    "os"
    "math/big"
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
    return viper.GetInt64("chainid")
}

func GetKeystore() string {
    return dataDir + string(os.PathSeparator) + "keystore"
}

func GetMyIndex() int {
    boot := IsBootNode()
    if boot {
        return viper.GetInt("myindex")
    } else {
        return -1
    }
}

func GetDebugNodes() []*DebugNode {
    config := &Config{}
    viper.Unmarshal(config)
    return config.DebugNodes
}

func GetPort() int {
    return viper.GetInt("port")
}

func GetBlockPrize() *big.Int {
    blockPrize := viper.GetString("blockprize")
    prize, _ := new(big.Int).SetString(blockPrize, 10)
    return prize
}

func IsBootNode() bool {
    return viper.GetBool("boot")
}