package abi

import (
	"github.com/drep-project/drep-chain/common"
	"math/big"
	"reflect"
)
// packElement packs the given reflect value according to the abi specification in
// t.
func packElement(sink common.ZeroCopySink ,t Type, reflectValue reflect.Value) {
	switch t.T {
	case int8Ty:
		sink.WriteInt8(int8(reflectValue.Int()))
	case int16Ty:
		sink.WriteInt16(int16(reflectValue.Int()))
	case int32Ty:
		sink.WriteInt32(int32(reflectValue.Int()))
	case int64Ty:
		sink.WriteInt64(int64(reflectValue.Int()))
	case uint8Ty:
		sink.WriteUint8(uint8(reflectValue.Int()))
	case uint16Ty:
		sink.WriteUint16(uint16(reflectValue.Int()))
	case uint32Ty:
		sink.WriteUint32(uint32(reflectValue.Int()))
	case uint64Ty:
		sink.WriteUint64(uint64(reflectValue.Int()))
	case u256Ty:
		var bigInt =  reflectValue.Interface().(*big.Int)
		sink.WriteU256(bigInt)
	case StringTy:
		sink.WriteString(reflectValue.String())
	case BoolTy:
		sink.WriteBool(reflectValue.Bool())
	case SliceTy:
		len := reflectValue.Len()
		sink.WriteVarUint(uint64(len))
		for i:=0;i<len;i++{
			packElement(sink, *t.Elem, reflectValue.Index(i))
		}
	case ArrayTy:
		len := reflectValue.Len()
		for i:=0;i<len;i++{
			packElement(sink, *t.Elem, reflectValue.Index(i))
		}
	default:
		panic("abi: fatal error")
	}
}

// packElement packs the given reflect value according to the abi specification in
// t.
func UnpackElement(reader *common.ZeroCopySource ,t Type) (interface{}, bool){
	switch t.T {
	case int8Ty:
		return reader.NextInt8()
	case int16Ty:
		return reader.NextInt16()
	case int32Ty:
		return reader.NextInt32()
	case int64Ty:
		return reader.NextInt64()
	case uint8Ty:
		return reader.NextUint8()
	case uint16Ty:
		return reader.NextUint16()
	case uint32Ty:
		return reader.NextUint32()
	case uint64Ty:
		return reader.NextUint64()
	case u256Ty:
		return reader.Nextu256()
	case StringTy:
		val,_,_,eof := reader.NextString()
		return val, eof
	case BoolTy:
		val,_,eof := reader.NextBool()
		return val, eof
	case SliceTy:
		len,_,_,_ := reader.NextVarUint()
		slice := reflect.MakeSlice(t.Elem.Type,int(len),int(len))
		for i:=0;i < int(len);i++{
			ele, _ := UnpackElement(reader, *t.Elem)
			slice.Index(i).Set(reflect.ValueOf(ele))
		}
		return slice, true
	case ArrayTy:
		val := reflect.New(t.Type).Elem()
		for i:=0;i<t.Size;i++{
			ele, _ := UnpackElement(reader, *t.Elem)
			val.Index(i).Set(reflect.ValueOf(ele))
		}
		return val, true
	default:
		panic("abi: fatal error")
	}
}

