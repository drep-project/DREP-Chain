package database

import (
	"os"
	"testing"
)

func TestNewDatabase(t *testing.T) {

	os.RemoveAll("./test")

	db, err := NewDatabase("./test")
	if err != nil {
		t.Fatal(err)
	}
	db.Close()

	_, err = NewDatabase("./test")
	if err == nil {
		t.Fatal(err)
	}

	os.RemoveAll("./test")
}

func TestAddLog(t *testing.T) {

}
