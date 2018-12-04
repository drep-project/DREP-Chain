package config

import (
    "testing"
    "github.com/spf13/viper"
    "fmt"
    "BlockChainTest/config/debug"
)

func TestConfigWrite(t *testing.T) {
    v := viper.New()
    v.SetConfigName("config")
    v.AddConfigPath(".")
    err := v.ReadInConfig()
    if err != nil {
       panic(fmt.Errorf("Fatal error config file: %s \n", err))
    }
    fmt.Println(v.Get("relay_node"))

    v.Set("ChainId", 30)
    v.Set("ABCdefg", "iuiu")
    v.Set("abcdefg", 34)
    v.WriteConfig()
}

func TestInit(t *testing.T) {
    //debug.Init()
}

func TestReadConfig(t *testing.T) {
    debug.GetDebugConfig(4)
}