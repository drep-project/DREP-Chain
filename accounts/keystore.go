package accounts

import (
    "os"
    "encoding/hex"
    "encoding/json"
    "path"
    "errors"
    "fmt"
)

var (
    KeystoreDirName = "keystore"
)

type Key struct {
    Address string
    PrivateKey string
    ChainId int64
    ChainCode string
}

func genKeystore(keyAddr string, jsonBytes []byte) error {
    os.MkdirAll(KeystoreDirName, os.ModeDir|os.ModePerm)
    filename := getFilename(keyAddr)
    file, err := os.Create(filename)
    if err != nil {
        return err
    }
    file.Write(jsonBytes)
    file.Close()
    return nil
}

func store(node *Node) error {
    key := &Key{
        Address: node.Address().Hex(),
        PrivateKey: hex.EncodeToString(node.PrvKey.Prv),
        ChainId: node.ChainId,
        ChainCode: hex.EncodeToString(node.ChainCode),
    }
    b, err := json.Marshal(key)
    if err != nil {
        return err
    }
    return genKeystore(key.Address, b)
}

func getFilename(keyAddr string) string {
    return path.Join(KeystoreDirName, keyAddr + ".json")
}

func OpenKeystore(keystorePath string) (*Node, error) {
    file, err := os.Open(keystorePath)
    defer file.Close()
    if err != nil {
        return nil, err
    }

    jsonBytes := make([]byte, 1024)
    n, err := file.Read(jsonBytes)
    if err != nil {
        return nil, err
    }

    jsonBytes = jsonBytes[:n]
    key := &Key{}
    err = json.Unmarshal(jsonBytes, key)
    if err != nil {
        return nil, err
    }

    prv, err := hex.DecodeString(key.PrivateKey)
    if err != nil {
        return nil, err
    }

    chainCode, err := hex.DecodeString(key.ChainCode)
    if err != nil {
        return nil, err
    }

    node := &Node{
        PrvKey:  genPrvKey(prv),
        ChainCode: chainCode,
    }
    return node, nil
}

func SaveKeystore(node *Node, keystorePath string) error {
    key := &Key{
        Address: node.Address().Hex(),
        PrivateKey: hex.EncodeToString(node.PrvKey.Prv),
        ChainId: node.ChainId,
        ChainCode: hex.EncodeToString(node.ChainCode),
    }
    b, err := json.Marshal(key)
    if err != nil {
        return err
    }

    if keystorePath == "" {
        panic("fuck")
        dataDir := ""//config.GetDataDir()
        fmt.Println("datadir: ", dataDir)
        if dataDir == "" {
            return errors.New("failed to get current data directory")
        }
        keystorePath = path.Join(dataDir, key.Address + ".json")
    }

    err = os.MkdirAll(KeystoreDirName, os.ModeDir|os.ModePerm)
    if err != nil {
        return err
    }
    file, err := os.Create(keystorePath)
    defer file.Close()
    if err != nil {
        return err
    }
    _, err = file.Write(b)
    if err != nil {
        return err
    }
    return nil
}

func MiniSave(node *Node) error {
    key := &Key{
        Address: node.Address().Hex(),
        PrivateKey: hex.EncodeToString(node.PrvKey.Prv),
        ChainId: node.ChainId,
        ChainCode: hex.EncodeToString(node.ChainCode),
    }
    b, err := json.Marshal(key)
    if err != nil {
        return err
    }
    file, err := os.Create(key.Address + ".json")
    defer file.Close()
    if err != nil {
        return err
    }
    _, err = file.Write(b)
    if err != nil {
        return err
    }
    return nil
}