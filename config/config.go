package config

import (
    "github.com/spf13/viper"
    "fmt"
    "sync"
)

const (
    RootChain = 0
)

var (
    DefaultDataDir = "data"
    DefaultKeystore = "keystore"
    DefaultChain = RootChain
    DefaultIP = ""
    DefaultPort = 55555

    MinerNum = 1
    MyIndex = 0

    once sync.Once
    vip *viper.Viper
)

type Config struct {
    RelayNode []string
    ChainId   int64
    DataDir   string
    Keystore  string
    IP        string
    Port      int
    MinerNum  int
    MyIndex   int
}

func init() {
    v := GetViper()
    v.Set("ChainId", DefaultChain)
    v.Set("DataDir", DefaultDataDir)
    v.Set("Keystore", DefaultKeystore)
    v.Set("IP", DefaultIP)
    v.Set("Port", DefaultPort)
    v.Set("MinerNum", MinerNum)
    v.Set("MyIndex", MyIndex)
    err := v.WriteConfig()
    fmt.Println("err: ", err)
}

func GetViper() *viper.Viper {
    once.Do(func() {
        if vip == nil {
            vip = viper.New()
            vip.SetConfigName("config")
            vip.AddConfigPath("./config")
        }
    })
    return vip
}

func GetConfig() (*Config, error) {
    v := GetViper()
    err := v.ReadInConfig()
    if err != nil {
        return nil, err
    }
    conf := &Config{}
    if err := v.Unmarshal(&conf) ; err != nil{
        return nil, err
    }
    return conf, nil
}

func GetChainId() int64 {
    conf, err := GetConfig()
    if err != nil {
        return RootChain
    }
    return conf.ChainId
}

func GetDataDir() string {
    return vip.GetString("datadir")
    //conf, err := GetConfig()
    //if err != nil {
    //    fmt.Println(5678, err)
    //    return ""
    //}
    //return conf.DataDir
}

func GetKeystore() string {
    conf, err := GetConfig()
    if err != nil {
        return ""
    }
    return conf.Keystore
}

func GetMinerNum() int {
    conf, err := GetConfig()
    if err != nil {
        return 0
    }
    return conf.MinerNum
}

func GetMyIndex() int {
    conf, err := GetConfig()
    if err != nil {
        return 0
    }
    return conf.MyIndex
}

func SetChain(chainId int64, dataDir string) error {
    v := GetViper()
    v.Set("ChainId", chainId)
    v.Set("DataDir", dataDir)
    return v.WriteConfig()
}

func SetKeystore(keystorePath string) error {
    v := GetViper()
    v.Set("Keystore", keystorePath)
    return v.WriteConfig()
}