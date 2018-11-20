package account

import (
    "os"
)

var (
    KeystoreDirName = "keystore"
)

type Key struct {
    PrivateKey string
    Address string
}

type KeyStore interface {
    IfExist(string) (bool, error)
    GetKey(string) (string, error)
    StoreKey(string) (string, error)
    JoinPath(filename string) string
}

func GenKeystore(keyAddr string, jsonBytes []byte) error {
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
    keystoreName := GetFilename(keyAddr)
    file, cfErr := os.Create(keystoreName)
    if cfErr != nil {
        return cfErr
    }
    file.Write(jsonBytes)
    return nil
}

func LoadKeystore(keyAddr string) ([]byte, error) {
    cdErr := os.Chdir(KeystoreDirName)
    if cdErr != nil {
        return nil, cdErr
    }
    keystoreName := GetFilename(keyAddr)
    file, opErr := os.Open(keystoreName)
    if opErr != nil {
        return nil, opErr
    }
    jsonObj := make([]byte, 1024)
    n, readErr := file.Read(jsonObj)
    return jsonObj[:n], readErr
}

func GetFilename(keyAddr string) string {
    return keyAddr
}

//func GetFilename(keyAddr string) string {
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