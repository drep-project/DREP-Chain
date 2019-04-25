package common

import (
	"math/big"
	"reflect"
	"github.com/drep-project/binary"
)

type commomBigCodeC struct{}

// Encode encodes a value into the encoder.
func (c *commomBigCodeC) EncodeTo(e *binary.Encoder, rv reflect.Value) error {
	fff :=  (big.Int)(rv.Interface().(Big))
	contents := fff.Bytes()
	e.WriteUvarint(uint64(len(contents)))
	e.Write(contents)
	return nil
}

// Decode decodes into a reflect value from the decoder.
func (c *commomBigCodeC) DecodeTo(d *binary.Decoder, rv reflect.Value) (err error) {
	len, err := d.ReadUvarint()
	if err != nil {
		return err
	}
	contents := make([]byte, len)
	_, err = d.Read(contents)
	if err != nil {
		return err
	}

	rv.Set(reflect.ValueOf((Big)(*new (big.Int).SetBytes(contents))))
	return nil
}

func init(){
	binary.ImportCodeC(reflect.TypeOf(Big{}), &commomBigCodeC{})
}