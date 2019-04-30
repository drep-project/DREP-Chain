package database

type journal struct {
	Op       string
	Key      []byte
	Value    []byte
	Previous []byte
}