package abi

import (
	"fmt"
	"reflect"
	"regexp"
	"strconv"
	"strings"
)

// Type enumerator
const (
	int8Ty byte = iota
	int16Ty
	int32Ty
	int64Ty
	uint8Ty
	uint16Ty
	uint32Ty
	uint64Ty
	u256Ty
	StringTy
	BoolTy
	SliceTy
	ArrayTy
)

// Type is the reflection of the supported argument type
type Type struct {
	Elem *Type

	Kind reflect.Kind
	Type reflect.Type
	Size int
	T    byte // Our own type checking

	stringKind string // holds the unparsed string for deriving signatures
}

// NewType creates a new reflection type of abi type given in t.
func NewType(t string) (typ Type, err error) {
	// check that array brackets are equal if they exist
	if strings.Count(t, "[") != strings.Count(t, "]") {
		return Type{}, fmt.Errorf("invalid arg type in abi")
	}

	typ.stringKind = t

	// if there are brackets, get ready to go into slice/array mode and
	// recursively create the type
	if strings.Count(t, "[") != 0 {
		i := strings.LastIndex(t, "[")
		// recursively embed the type
		embeddedType, err := NewType(t[:i])
		if err != nil {
			return Type{}, err
		}
		// grab the last cell and create a type from there
		sliced := t[i:]
		// grab the slice size with regexp
		re := regexp.MustCompile("[0-9]+")
		intz := re.FindAllString(sliced, -1)

		if len(intz) == 0 {
			// is a slice
			typ.T = SliceTy
			typ.Kind = reflect.Slice
			typ.Elem = &embeddedType
			typ.Type = reflect.SliceOf(embeddedType.Type)
		} else if len(intz) == 1 {
			// is a array
			typ.T = ArrayTy
			typ.Kind = reflect.Array
			typ.Elem = &embeddedType
			typ.Size, err = strconv.Atoi(intz[0])
			if err != nil {
				return Type{}, fmt.Errorf("abi: error parsing variable size: %v", err)
			}
			typ.Type = reflect.ArrayOf(typ.Size, embeddedType.Type)
		} else {
			return Type{}, fmt.Errorf("invalid formatting of array type")
		}
		return typ, err
	}
	// varType is the parsed abi type
	switch t {
	case "int8":
		typ.Kind = reflect.Int8
		typ.Type =  int8T
		typ.Size = 1
		typ.T = int8Ty
	case "int16":
		typ.Kind = reflect.Int16
		typ.Type =  int16T
		typ.Size = 2
		typ.T = int16Ty
	case "int32":
		typ.Kind = reflect.Int32
		typ.Type =  int32T
		typ.Size = 4
		typ.T = int32Ty
	case "int64":
		typ.Kind = reflect.Int64
		typ.Type =  int64T
		typ.Size = 8
		typ.T = int64Ty
	case "uint8":
		typ.Kind = reflect.Uint8
		typ.Type =  uint8T
		typ.Size = 1
		typ.T = uint8Ty
	case "uint16":
		typ.Kind = reflect.Uint16
		typ.Type =  uint16T
		typ.Size = 2
		typ.T = uint16Ty
	case "uint32":
		typ.Kind = reflect.Uint32
		typ.Type =  uint32T
		typ.Size = 4
		typ.T = uint32Ty
	case "uint64":
		typ.Kind = reflect.Uint64
		typ.Type =  uint64T
		typ.Size = 8
		typ.T = uint64Ty
	case "u256":
	//	typ.Kind = reflect.u2
	//	typ.Type =  uint64T
		typ.Size = 32
		typ.T = u256Ty
	case "bool":
		typ.Kind = reflect.Bool
		typ.T = BoolTy
		typ.Type = reflect.TypeOf(bool(false))
	case "string":
		typ.Kind = reflect.String
		typ.Type = reflect.TypeOf("")
		typ.T = StringTy
	default:
		return Type{}, fmt.Errorf("unsupported arg type: %s", t)
	}

	return
}