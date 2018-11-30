package config

import (
    "github.com/spf13/viper"
    "sync"
)

const (
    RootChain = 0
)

var (
    once sync.Once
    singleton *Config
)

type Config struct {
    RelayNode []string
    ChainId   int64
    DataDir   string
    Keystore  string
}

func GetConfigInstance() *Config {
    once.Do(func() {
        singleton = &Config{}
    })
    return singleton
}

func GetViper() *viper.Viper {
    v := viper.New()
    v.SetConfigName("config")
    v.AddConfigPath(".")
    return v
}

func GetConfig() (*Config, error) {
    v := GetViper()
    err := v.ReadInConfig()
    if err != nil {
        return nil, err
    }
    conf := GetConfigInstance()
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
    conf, err := GetConfig()
    if err != nil {
        return ""
    }
    return conf.DataDir
}

func GetKeystore() string {
    conf, err := GetConfig()
    if err != nil {
        return ""
    }
    return conf.Keystore
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