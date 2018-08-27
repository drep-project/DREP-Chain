package database
//
//import (
//  "github.com/syndtr/goleveldb/leveldb"
//  "github.com/syndtr/goleveldb/leveldb/iterator"
//)
//
//type Database struct {
//  LevelDB *leveldb.DB
//  FilePath string
//}
//
//func NewDatabase(filePath string) (*Database, error) {
// ldb, err := leveldb.OpenFile(filePath, nil)
// return &Database{ldb, filePath}, err
//}
//
////func LocalKey(key string, value interface{}) (string, error) {
////    switch value.(type) {
////    case string:
////        return IP_KEY_SUFFIX + key, nil
////    case *common.Block:
////        return BLOCK_KEY_SUFFIX + key, nil
////    case *common.Transaction:
////        return TRANSACTION_KEY_SUFFIX + key, nil
////    default:
////        return "", errors.New("unknown value type")
////    }
////}
//
//func (db *Database) Get(key string) (interface{}, error) {
//  b, err := db.LevelDB.Get([]byte(key), nil)
//  if err != nil {
//      return nil, err
//  }
//  value, err := common.Deserialize(b)
//  if err != nil {
//      return nil, err
//  }
//  return value, nil
//}
//
//func (db *Database) Put(key string, value interface{}) error {
//  b, err := common.Serialize(value)
//  if err != nil {
//      return err
//  }
//  return db.LevelDB.Put([]byte(key), b, nil)
//}
//
//func (db *Database) Delete(key string) error {
//  return db.LevelDB.Delete([]byte(key), nil)
//}
//
//func (db *Database) Close() error {
//  return db.LevelDB.Close()
//}
//
//type Iterator struct {
//  Itr iterator.Iterator
//}
//
//func (db *Database) NewIterator() *Iterator {
//  return &Iterator{db.LevelDB.NewIterator(nil, nil)}
//}
//
//func (itr *Iterator) Next() bool {
//  return itr.Itr.Next()
//}
//
//func (itr *Iterator) Key() string {
//  return string(itr.Key())
//}
//
//func (itr *Iterator) Value() interface{} {
//  value, _ := common.Deserialize(itr.Itr.Value())
//  return value
//}