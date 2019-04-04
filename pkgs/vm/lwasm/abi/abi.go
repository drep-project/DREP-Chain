package abi

import (
	"encoding/json"
	"errors"
	"github.com/drep-project/drep-chain/common"
	"reflect"
)

// The ABI holds information about a contract's context and available
// invokable methods. It will allow you to type check function calls and
// packs data accordingly.
type ABI struct {
	Name        string
	Methods     map[string]Method
	Events      map[string]Event
}

func (abi ABI) PackMethod(name string, args ...interface{}) ([]byte, error) {
	sink := common.ZeroCopySink{}
	method, ok := abi.Methods[name];
	if  !ok {
		return nil, errors.New("method not exist")
	}
	if len(method.Inputs) != len(args) {
		return nil, errors.New("args num not correct")
	}
	sink.WriteString(name)
	for i:=0; i< len(args);i++{
		packElement(sink,method.Inputs[i].Type,reflect.ValueOf(args[i]))
	}
	return sink.Bytes(), nil
}

func (abi ABI) UnPackMethod(name string, output []byte) (interface{}, error) {
	method, ok := abi.Methods[name];
	if  !ok {
		return nil, errors.New("method not exist")
	}
	reader := common.NewZeroCopySource(output)
	val, _ := UnpackElement(reader,method.Output)
	return val, nil
}

// Unpack output in v according to the abi specification
func (abi ABI) UnPackEvent(name string, output []byte) ([]interface{}, error) {
	event, ok := abi.Events[name];
	if  !ok {
		return nil, errors.New("method not exist")
	}
	vals := []interface{}{}
	reader := common.NewZeroCopySource(output)
	for _, ele := range event.Inputs {
		val, _ := UnpackElement(reader, ele.Type)
		vals = append(vals, val)
	}
	return vals, nil
}

// UnmarshalJSON implements json.Unmarshaler interface
func (abi *ABI) UnmarshalJSON(data []byte) error {
	abitemp := &ABI{}
	err := json.Unmarshal(data, abitemp)
	if err != nil {
		return err
	}
	abi.Name = abitemp.Name
	abi.Methods =  abitemp.Methods
	abi.Events = abitemp.Events

	return nil
}
