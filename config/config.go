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

func GetDb() string {
    return dataDir + string(os.PathSeparator) + "database"
}

func GetMyIndex() int {
    boot := IsBootNode()
    if boot {
        return viper.GetInt("myindex")
    } else {
        return -1
    }
}

func GetDebugNodes() []*BootNode {
    config := &struct {
        BootNodes []*BootNode
    }{}
    viper.Unmarshal(config)
    return config.BootNodes
}

func GetPort() int {
    port := viper.GetInt("port")
    if port == 0 {
        return defaultPort
    } else {
        return port
    }
}

func GetBlockPrize() *big.Int {
    blockPrize := viper.GetString("blockprize")
    if blockPrize == "" {
        blockPrize = defaultBlockPrize
    }
    prize, _ := new(big.Int).SetString(blockPrize, 10)
    return prize
}

func IsBootNode() bool {
    return viper.GetBool("boot")
}