package codec

import (
	"bytes"
	"encoding/binary"
	"errors"
	"fmt"
	"reflect"
	"time"
	"unsafe"
)

// Optimizations:
// For string/slice/array, length of data dont need to be 64 bits

// Add-on Kinds:
// Time = 27

type Encoder struct{}

func (enc *Encoder) Marshall(data interface{}) ([]byte, error) {
	if data == nil {
		return []byte{}, nil
	}

	rv := reflect.ValueOf(data)
	switch rv.Kind() {
	case reflect.Ptr:
		return enc.marshallPtr(rv.Interface())
	case reflect.Struct:
		if rv.Type() == reflect.TypeOf(time.Now()) {
			return enc.marshallTime(rv.Interface())
		}
		return enc.marshallStruct(rv.Interface())
	case reflect.Map:
		return enc.marshallMap(rv.Interface())
	case reflect.Array, reflect.Slice:
		return enc.marshallIterable(rv.Interface())
	case reflect.String:
		return enc.marshallString(rv.Interface())
	case reflect.Bool,
		reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64,
		reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64,
		reflect.Float32, reflect.Float64,
		reflect.Complex64, reflect.Complex128:
		return enc.marshallNumber(rv.Interface())
	case reflect.Invalid,
		reflect.Uintptr, reflect.UnsafePointer,
		reflect.Chan, reflect.Func:
		return []byte{}, fmt.Errorf("unable to encode kind %d", rv.Kind())
	default:
		return []byte{}, nil
	}
}

func (enc *Encoder) marshallPtr(data interface{}) ([]byte, error) {
	rv := reflect.ValueOf(data)
	dereferencedRv := rv.Elem()

	dereferencedBytes, err := enc.Marshall(dereferencedRv.Interface())
	if err != nil {
		return []byte{}, err
	}

	// kindByte := byte(reflect.Ptr)
	result := []byte{}
	result = append(result, dereferencedBytes...)
	return result, nil
}

func (enc *Encoder) marshallTime(data interface{}) ([]byte, error) {
	// Format: [kind (8-bit)][time as int]
	timeData, ok := data.(time.Time)
	if !ok {
		return []byte{}, fmt.Errorf("invalid time: %s", timeData)
	}

	kindByte := byte(27)
	timeStrBytes, err := enc.marshallNumber(timeData.UnixMilli())
	if err != nil {
		return []byte{}, err
	}

	return append([]byte{kindByte}, timeStrBytes...), nil
}

func (enc *Encoder) marshallIterable(data interface{}) ([]byte, error) {
	// Format: [kind (8-bit)][number (64-bits) of items][iterable items]
	rv := reflect.ValueOf(data)
	typeByte := byte(rv.Kind())
	numItemsBytes, err := enc.numToBytes(rv.Len())
	if err != nil {
		return []byte{}, err
	}

	result := append([]byte{}, typeByte)
	result = append(result, numItemsBytes...)

	for i := 0; i < rv.Len(); i++ {
		item := rv.Index(i).Interface()
		itemBytes, err := enc.Marshall(item)
		if err != nil {
			return []byte{}, err
		}
		result = append(result, itemBytes...)
	}

	return result, nil
}

func (enc *Encoder) marshallMap(data interface{}) ([]byte, error) {
	// Format: [kind (8-bit)][number (64-bits) of kv pairs][kv pairs]
	rv := reflect.ValueOf(data)
	typeByte := byte(rv.Kind())
	numKeyValuePairsBytes, err := enc.numToBytes(rv.Len())
	if err != nil {
		return []byte{}, err
	}
	result := append([]byte{}, typeByte)
	result = append(result, numKeyValuePairsBytes...)

	for _, key := range rv.MapKeys() {
		var keyBytes []byte
		if key.Kind() == reflect.Interface {
			keyBytes, err = enc.marshallInterface(key.Interface())
			if err != nil {
				return []byte{}, err
			}
		} else {
			keyBytes, err = enc.Marshall(key.Interface())
			if err != nil {
				return []byte{}, err
			}
		}

		// keyBytes, err := enc.Marshall(key.Interface())
		// if err != nil {
		// 	return []byte{}, err
		// }

		value := rv.MapIndex(key)

		var valueBytes []byte
		if value.Kind() == reflect.Interface {
			valueBytes, err = enc.marshallInterface(value.Interface())
			if err != nil {
				return []byte{}, err
			}
		} else {
			valueBytes, err = enc.Marshall(value.Interface())
			if err != nil {
				return []byte{}, err
			}
		}

		result = append(result, keyBytes...)
		result = append(result, valueBytes...)
	}
	return result, nil
}

func (enc *Encoder) marshallStruct(data interface{}) ([]byte, error) {
	// Format: [kind (8-bit)][number (8-bit) of fields][field-value pairs]
	// TODO: Assuming structs usually have < 255 fields?

	rv := reflect.ValueOf(data)
	kindByte := byte(rv.Kind())
	numFieldsByte := byte(rv.NumField())
	result := append([]byte{}, kindByte, numFieldsByte)

	structType := rv.Type()
	for i := 0; i < rv.NumField(); i++ {
		fieldKey := structType.Field(i)
		keyBytes, err := enc.Marshall(fieldKey.Name)
		if err != nil {
			return []byte{}, err
		}

		fieldValue := rv.Field(i)

		var valueBytes []byte
		if fieldValue.Kind() == reflect.Interface {
			valueBytes, err = enc.marshallInterface(fieldValue.Interface())
			if err != nil {
				return []byte{}, err
			}
		} else {
			valueBytes, err = enc.Marshall(fieldValue.Interface())
			if err != nil {
				return []byte{}, err
			}
		}

		result = append(result, keyBytes...)
		result = append(result, valueBytes...)
	}
	return result, nil
}

func (enc *Encoder) marshallInterface(data interface{}) ([]byte, error) {
	// Format: [kind (8-bit)][total RHS length (64-bit)][value length (64-bit)][value]
	// Saves the interface data as raw byte slice to be handled by Decoder

	kindByte := byte(reflect.Interface)
	valueBytes, err := enc.Marshall(data)
	if err != nil {
		return []byte{}, err
	}

	dataLenBytes, err := enc.numToBytes(len(valueBytes))
	if err != nil {
		return []byte{}, err
	}

	result := append([]byte{}, kindByte)
	result = append(result, dataLenBytes...)
	result = append(result, valueBytes...)
	return result, nil
}

func (enc *Encoder) marshallString(data interface{}) ([]byte, error) {
	// Format: [kind (8-bit)][value length (64-bit)][value]
	// Convert aliases to the underlying primitive
	data = reflect.ValueOf(data).Convert(kindToType[reflect.String]).Interface()
	kindByte := byte(reflect.String)
	dataStr, ok := data.(string)
	if !ok {
		return []byte{}, errors.New("uncastable to string")
	}

	dataBytes, err := enc.strToBytes(dataStr)
	if err != nil {
		return []byte{}, err
	}

	dataLenBytes, err := enc.numToBytes(len(dataBytes))
	if err != nil {
		return []byte{}, err
	}

	result := append([]byte{}, kindByte)
	result = append(result, dataLenBytes...)
	result = append(result, dataBytes...)
	return result, nil
}

func (enc *Encoder) marshallNumber(data interface{}) ([]byte, error) {
	// Format: [kind (8-bit)][length (8-bit)][value]
	// Convert aliases to the underlying primitive
	data = reflect.ValueOf(data).Convert(kindToType[reflect.TypeOf(data).Kind()]).Interface()
	kindByte := byte(reflect.ValueOf(data).Kind())
	dataBytes, err := enc.numToBytes(data)
	if err != nil {
		return []byte{}, err
	}
	numBytes := byte(len(dataBytes)) // Max Value = 8
	result := append([]byte{}, kindByte, numBytes)
	result = append(result, dataBytes...)
	return result, nil
}

func (enc *Encoder) strToBytes(data string) ([]byte, error) {
	res := make([]byte, 0, 4*len(data)) // char == rune == int32
	for idx := 0; idx < len(data); idx++ {
		numAsBytes, err := enc.numToBytes(data[idx])
		if err != nil {
			return []byte{}, err
		}
		res = append(res, numAsBytes...)
	}
	return res, nil
}

func (enc *Encoder) numToBytes(data interface{}) ([]byte, error) {
	// Cast to exact types for binary.Write
	var castedData interface{}
	switch v := data.(type) {
	case int:
		castedData = int64(v)
	case uint:
		castedData = uint64(v)
	default:
		castedData = data
	}

	buf := bytes.NewBuffer(make([]byte, 0))
	err := binary.Write(buf, endianness, castedData)
	if err != nil {
		return []byte{}, err
	}

	return buf.Bytes(), nil
}

var kindToType = map[reflect.Kind]reflect.Type{
	reflect.Bool:          reflect.TypeOf(false),
	reflect.Int:           reflect.TypeOf(int(0)),
	reflect.Int8:          reflect.TypeOf(int8(0)),
	reflect.Int16:         reflect.TypeOf(int16(0)),
	reflect.Int32:         reflect.TypeOf(int32(0)),
	reflect.Int64:         reflect.TypeOf(int64(0)),
	reflect.Uint:          reflect.TypeOf(uint(0)),
	reflect.Uint8:         reflect.TypeOf(uint8(0)),
	reflect.Uint16:        reflect.TypeOf(uint16(0)),
	reflect.Uint32:        reflect.TypeOf(uint32(0)),
	reflect.Uint64:        reflect.TypeOf(uint64(0)),
	reflect.Uintptr:       reflect.TypeOf(uintptr(0)),
	reflect.Float32:       reflect.TypeOf(float32(0)),
	reflect.Float64:       reflect.TypeOf(float64(0)),
	reflect.Complex64:     reflect.TypeOf(complex64(0)),
	reflect.Complex128:    reflect.TypeOf(complex128(0)),
	reflect.String:        reflect.TypeOf(""),
	reflect.UnsafePointer: reflect.TypeOf((unsafe.Pointer(nil))),
}

// func (c *Codec) marshallPrimitive(data interface{}) ([]byte, error) {
// 	// Format: <type><value length; 4 bytes><value>
// 	rv := reflect.ValueOf(data)
// 	typeByte := byte(rv.Kind()) // Guaranteed 1 <= x <= 26 (fits within a byte)

// 	valueBytes, err := enc.primitiveToBytes(data)
// 	if err != nil {
// 		return []byte{}, err
// 	}
// 	valueLenBytes := enc.numToBytes(len(valueBytes))

// 	result := append([]byte{}, typeByte)
// 	result = append(result, valueLenBytes...)
// 	result = append(result, valueBytes...)
// 	return result, nil
// }

// func (c *Codec) primitiveToBytes(data interface{}) ([]byte, error) {
// 	// TODO Have to manually do this?
// 	buf := new(bytes.Buffer)
// 	enc := gob.NewEncoder(buf)
// 	if err := enenc.Encode(data); err != nil {
// 		return []byte{}, err
// 	}
// }

// 	return buf.Bytes(), nil
