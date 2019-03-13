package types

import (
	"encoding/json"
	"errors"
	"log"
	"testing"
)

func TestTransferActionMashalAndUnmarshal(t *testing.T) {
	action := TransferAction{
		To:"jimmy",
	}
	bytes, err := json.Marshal(action)
	if err != nil {
		log.Fatal(err)
	}

	newAction := &TransferAction{}
	err = json.Unmarshal(bytes, newAction)
	if err != nil {
		log.Fatal(err)
	}

	if action.To != newAction.To {
		log.Fatal(errors.New("not matched"))
	}
}