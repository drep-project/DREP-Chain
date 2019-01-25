package config

import (
    "github.com/spf13/viper"
    "fmt"
    "os"
    "math/big"
    "strconv"
)

const (
    RootChain = 0
)

var (
    dataDir = "." + string(os.PathSeparator) + "data"
    viper0 = viper.New()
    viper1 = viper.New()
)

func init() {
    viper.SetConfigName("config")
    viper.AddConfigPath(dataDir)
    err := viper.ReadInConfig()
    if err != nil {
        fmt.Println("read config file error: ", err)
    }

    viper0.SetConfigName("rep_configs")
    viper0.AddConfigPath(dataDir)
    err = viper0.ReadInConfig()
    if err != nil {
        fmt.Println("read rep config file error: ", err)
    }

    viper1.SetConfigName("rep_floats")
    viper1.AddConfigPath(dataDir)
    err = viper1.ReadInConfig()
    if err != nil {
        fmt.Println("read rep floats file error: ", err)
    }
    fmt.Println("viper keys: ",viper.AllKeys())
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

func GetAdminNodes() []*AdminNode {
    config := &struct {
        AdminNodes []*AdminNode
    }{}
    viper.Unmarshal(config)
    return config.AdminNodes
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

func GetBottomRate() *big.Float {
    br := viper0.GetString("bottomrate")
    if br == "" {
        return nil
    }
    bottomRate, ok := new(big.Float).SetString(br)
    if !ok {
        return nil
    }
    return bottomRate
}

func GetCutOff() *big.Float {
    co := viper0.GetString("cutoff")
    if co == "" {
        return nil
    }
    cutOff, ok := new(big.Float).SetString(co)
    if !ok {
        return nil
    }
    return cutOff
}

func GetEpsilon() *big.Float {
    eps := viper0.GetString("epsilon")
    if eps == "" {
        return nil
    }
    epsilon, ok := new(big.Float).SetString(eps)
    if !ok {
        return nil
    }
    return epsilon
}

func GetBeta() *big.Float {
    b := viper0.GetString("beta")
    if b == "" {
        return nil
    }
    beta, ok := new(big.Float).SetString(b)
    if !ok {
        return nil
    }
    return beta
}

func GetRequire() *big.Float {
    req := viper0.GetString("require")
    if req == "" {
        return nil
    }
    require, ok := new(big.Float).SetString(req)
    if !ok {
        return nil
    }
    return require
}

func getT(i int) *big.Float {
    t := viper1.GetString("t" + strconv.Itoa(i))
    if t == "" {
        return nil
    }
    tf, ok := new(big.Float).SetString(t)
    if !ok {
        return nil
    }
    return tf
}

func getE(i int) *big.Float {
    e := viper1.GetString("e" + strconv.Itoa(i))
    if e == "" {
        return nil
    }
    ef, ok := new(big.Float).SetString(e)
    if !ok {
        return nil
    }
    return ef
}

func getR(i int) *big.Float {
    r := viper1.GetString("r" + strconv.Itoa(i))
    if r == "" {
        return nil
    }
    rf, ok := new(big.Float).SetString(r)
    if !ok {
        return nil
    }
    return rf
}

func getAlpha1(i int) *big.Float {
    alpha := viper1.GetString("alpha1_" + strconv.Itoa(i))
    if alpha == "" {
        return nil
    }
    alpha1, ok := new(big.Float).SetString(alpha)
    if !ok {
        return nil
    }
    return alpha1
}

func getAlpha2(i, index int) *big.Float {
    alpha := viper1.GetString("alpha2_" + strconv.Itoa(i) + "_" + strconv.Itoa(index))
    if alpha == "" {
        return nil
    }
    alpha2, ok := new(big.Float).SetString(alpha)
    if !ok {
        return nil
    }
    return alpha2
}

func GetLinearPeriod() *big.Float {
    mode := viper0.GetString("decaymode")
    switch mode {
    case "faster":
        return getT(1)
    case "fast":
        return getT(2)
    case "medium":
        return getT(3)
    case "slow":
        return getT(4)
    case "slower":
        return getT(5)
    default:
        return nil
    }
}

func GetExponentialPeriod() *big.Float {
    mode := viper0.GetString("decaymode")
    switch mode {
    case "faster":
        return getE(1)
    case "fast":
        return getE(2)
    case "medium":
        return getE(3)
    case "slow":
        return getE(4)
    case "slower":
        return getE(5)
    default:
        return nil
    }
}

func GetAttenuationRate() *big.Float {
    mode := viper0.GetString("decaymode")
    switch mode {
    case "faster":
        return getR(1)
    case "fast":
        return getR(2)
    case "medium":
        return getR(3)
    case "slow":
        return getR(4)
    case "slower":
        return getR(5)
    default:
        return nil
    }
}

func GetAlpha1() *big.Float {
    mode := viper0.GetString("decaymode")
    switch mode {
    case "faster":
        return getAlpha1(1)
    case "fast":
        return getAlpha1(2)
    case "medium":
        return getAlpha1(3)
    case "slow":
        return getAlpha1(4)
    case "slower":
        return getAlpha1(5)
    default:
        return nil
    }
}

func GetAlpha2(index int) *big.Float {
    mode := viper0.GetString("decaymode")
    switch mode {
    case "faster":
        return getAlpha2(1, index)
    case "fast":
        return getAlpha2(2, index)
    case "medium":
        return getAlpha2(3, index)
    case "slow":
        return getAlpha2(4, index)
    case "slower":
        return getAlpha2(5, index)
    default:
        return nil
    }
}