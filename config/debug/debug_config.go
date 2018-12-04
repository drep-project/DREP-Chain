package debug

import (
    "github.com/spf13/viper"
    "strconv"
    "path"
    "os"
    "BlockChainTest/mycrypto"
    "encoding/hex"
    "math/big"
    "fmt"
)

const (
    DebugMode2 = 2
    DebugMode3 = 3
    DebugMode4 = 4
)

var (
    prv0 = []byte{0x22, 0x11}
    prv1 = []byte{0x14, 0x44}
    prv2 = []byte{0x11, 0x55}
    prv3 = []byte{0x31, 0x63}

    Nodes = []*DebugNode{
        {
            Index:   0,
            Prv:     new(big.Int).SetBytes(prv0).String(),
            PubKey:  FormatPubKey(mycrypto.GetCurve().ScalarBaseMultiply(prv0)),
            Address: "c6196f8d8165c7cbb5ffc3833d4caf0c92017c5d",
            IP:      "192.168.3.231",
            Port:    55555,
        },
        {
            Index:   1,
            Prv:     new(big.Int).SetBytes(prv1).String(),
            PubKey:  FormatPubKey(mycrypto.GetCurve().ScalarBaseMultiply(prv1)),
            Address: "aa5553c8da80f18e39ee0c5496d4fb2fdef075d3",
            IP:      "192.168.3.197",
            Port:    55555,
        },
        {
            Index:   2,
            Prv:     new(big.Int).SetBytes(prv2).String(),
            PubKey:  FormatPubKey(mycrypto.GetCurve().ScalarBaseMultiply(prv2)),
            Address: "34fa3eaa7fbff1ca46ca08a938835245e2f20a12",
            IP:      "192.168.3.236",
            Port:    55555,
        },
        {
            Index:   3,
            Prv:     new(big.Int).SetBytes(prv3).String(),
            PubKey:  FormatPubKey(mycrypto.GetCurve().ScalarBaseMultiply(prv3)),
            Address: "1f652d162f08f4db09e0aba0078fed4b8779cebe",
            IP:      "192.168.3.xxx",
            Port:    55555,
        },
    }
)

type PK struct {
    X string
    Y string
}

func FormatPubKey(pubKey *mycrypto.Point) *PK {
    return &PK{
        X: hex.EncodeToString(pubKey.X),
        Y: hex.EncodeToString(pubKey.Y),
    }
}

func ParsePK(pk *PK) *mycrypto.Point {
    x, _ := hex.DecodeString(pk.X)
    y, _ := hex.DecodeString(pk.Y)
    return &mycrypto.Point{
        X: x,
        Y: y,
    }
}

type DebugNode struct {
    Index   int
    Prv     string
    PubKey  *PK
    Address string
    IP      string
    Port    int
}

type DebugInfo struct {
    MinerNum   int
    MyIndex    int
    DebugNodes []*DebugNode
}

func init1() {
    for minerNum := 1; minerNum < 5; minerNum++ {
        configName := "config_" + strconv.FormatInt(int64(minerNum), 10) + "_nodes"
        if minerNum == 1 {
            configName = configName[:len(configName) - 1]
        }
        configPath := path.Join("debug", configName + ".json")
        file, err := os.Open(configPath)
        if err != nil {
            file, err = os.Create(configPath)
            InitDebugConfig(configName, minerNum)
        }
        file.Close()
    }
}

func InitDebugConfig(configName string, minerNum int) {
    v := viper.New()
    v.SetConfigName(configName)
    v.AddConfigPath("./config/debug")
    v.Set("MinerNum", minerNum)
    v.Set("MyIndex", -1)
    v.Set("DebugNodes", Nodes)
    v.WriteConfig()
}

func GetDebugConfig(minerNum int) *DebugInfo {
    configName := "config_" + strconv.FormatInt(int64(minerNum), 10) + "_nodes"
    if minerNum == 1 {
        configName = configName[:len(configName) - 1]
    }
    v := viper.New()
    v.SetConfigName(configName)
    v.AddConfigPath("./config/debug")
    err := v.ReadInConfig()
    if err != nil {
        fmt.Println("read err: ", err)
        return nil
    }
    fmt.Println(v.Get("debugnodes"))
    deb := &DebugInfo{}
    if err := v.Unmarshal(&deb) ; err != nil{
        fmt.Println("unm err: ", err)
        return nil
    }
    fmt.Println("deb: ", deb)
    return deb
}
