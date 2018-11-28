package accounts

import (
    "os"
    "encoding/hex"
    "encoding/json"
    "path"
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
    os.Mkdir(KeystoreDirName, os.ModeDir|os.ModePerm)
    filename := getFilename(keyAddr)
    file, err := os.Create(filename)
    if err != nil {
        return err
    }
    file.Write(jsonBytes)
    file.Close()
    return nil
}

func load(keyAddr string) ([]byte, error) {
    filename := getFilename(keyAddr)
    file, err := os.Open(filename)
    if err != nil {
        return nil, err
    }
    jsonObj := make([]byte, 1024)
    n, err := file.Read(jsonObj)
    file.Close()
    return jsonObj[:n], err
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
    return path.Join(KeystoreDirName, keyAddr)
}

func OpenKeystore(addr string) (*Node, error) {
    jsonBytes, err := load(addr)
    if err != nil {
        return nil, err
    }
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

//func getFilename(keyAddr string) string {
//    ts := time.Now().UTC()
//    return fmt.Sprintf("UTC--%s--%s", toISO8601(ts), keyAddr)
//}
//
//func toISO8601(t time.Time) string {
//    var timeZone string
//    name, offset := t.Zone()
//    if name == "UTC" {
//        timeZone = "Z"
//    } else {
//        timeZone = fmt.Sprintf("%03d00", offset/3600)
//    }
//    return fmt.Sprintf("%04d-%02d-%02dT%02d-%02d-%02d.%09d%s", t.Year(), t.Month(), t.Day(), t.Hour(), t.Minute(), t.Second(), t.Nanosecond(), timeZone)
//
//}