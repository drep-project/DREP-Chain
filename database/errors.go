package database

import (
	"errors"
)

var (
	ErrNoStorage       = errors.New("no account storage found")
	ErrKeyNotFound     = errors.New("key not found")
	ErrKeyUnSpport     = errors.New("unsupport")
	ErrUsedAlias       = errors.New("the alias has been used")
	ErrInvalidateAlias = errors.New("set null string as alias")
)
