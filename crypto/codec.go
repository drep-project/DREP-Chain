package crypto

import (
	"github.com/drep-project/binary"
	"github.com/drep-project/DREP-Chain/crypto/secp256k1"
	"reflect"
)

func init() {
	binary.ImportCodeC(reflect.TypeOf(secp256k1.PublicKey{}), &secpPubKeyCodeC{})
}

type secpPubKeyCodeC struct{}

// Encode encodes a value into the encoder.
func (c *secpPubKeyCodeC) EncodeTo(e *binary.Encoder, rv reflect.Value) error {
	pk := rv.Interface().(secp256k1.PublicKey)
	contents := pk.SerializeCompressed()
	e.WriteUvarint(uint64(len(contents)))
	e.Write(contents)
	return nil
}

// Decode decodes into a reflect value from the decoder.
func (c *secpPubKeyCodeC) DecodeTo(d *binary.Decoder, rv reflect.Value) (err error) {
	length, err := d.ReadUvarint()
	if err != nil {
		return err
	}
	contents := make([]byte, length)
	_, err = d.Read(contents)
	if err != nil {
		return err
	}
	pk, err := secp256k1.ParsePubKey(contents)
	if err != nil {
		return err
	}
	rv.Set(reflect.ValueOf(*pk))
	return nil
}
