package types

import (
	"encoding/json"
	"fmt"
	"math/big"
)

type Blockprize struct {
	big.Int
}

func (ip *Blockprize) MarshalJSON() ([]byte, error) {
	res := fmt.Sprintf("%x", ip)
	return json.Marshal(res)
}

func (ip *Blockprize) UnmarshalJSON(b []byte) error {
	var val string
	err := json.Unmarshal(b, &val)
	if err != nil {
		panic(err)
	}
	ip.SetString(val, 10)
	return nil
}