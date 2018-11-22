package accounts

import (
    "os"
    "encoding/hex"
    "encoding/json"
)

var (
    KeystoreDirName = "keystore"
)

type Key struct {
    ChainId int64
    PrivateKey string
    Address string
}

func genKeystore(keyAddr string, jsonBytes []byte) error {
    cdErr := os.Chdir(KeystoreDirName)
    if cdErr != nil {
        if os.IsNotExist(cdErr) {
            if mkdErr := os.Mkdir(KeystoreDirName, os.ModeDir); mkdErr != nil {
                return mkdErr
            }
        } else {
            return cdErr
        }
    }
    keystoreName := getFilename(keyAddr)
    file, cfErr := os.Create(keystoreName)
    if cfErr != nil {
        return cfErr
    }
    file.Write(jsonBytes)
    return nil
}

func load(keyAddr string) ([]byte, error) {
    cdErr := os.Chdir(KeystoreDirName)
    if cdErr != nil {
        return nil, cdErr
    }
    keystoreName := getFilename(keyAddr)
    file, opErr := os.Open(keystoreName)
    if opErr != nil {
        return nil, opErr
    }
    jsonObj := make([]byte, 1024)
    n, readErr := file.Read(jsonObj)
    return jsonObj[:n], readErr
}

func store(node *Node) error {
    key := &Key{
        ChainId: int64(node.ChainId),
        PrivateKey: hex.EncodeToString(node.PrvKey.Prv),
        Address: node.Address.Hex(),
    }
    b, err := json.Marshal(key)
    if err != nil {
        return err
    }
    return genKeystore(key.Address, b)
}

func getFilename(keyAddr string) string {
    return keyAddr
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
    node := &Node{
        ChainId: ChainID(key.ChainId),
        PrvKey:  genPrvKey(prv),
        Address: Hex2Address(key.Address),
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