package codec

import (
	"bytes"
	"encoding/binary"
	"errors"
	"fmt"
	"reflect"
	"time"
)

// Optimizations:

// Add-on Kinds:

type Encoder struct{}

func (enc *Encoder) Marshall(data interface{}) ([]byte, error) {
	rv := reflect.ValueOf(data)
	switch rv.Kind() {
	case reflect.Ptr:
		return enc.marshallPtr(data)
	case reflect.Struct:
		if rv.Type() == reflect.TypeOf(time.Now()) {
			return enc.marshallTime(data)
		}
		return enc.marshallStruct(data)
	case reflect.Map:
		return enc.marshallMap(data)
	case reflect.Array, reflect.Slice:
		return enc.marshallIterable(data)
	case reflect.String:
		return enc.marshallString(data)
	case reflect.Bool,
		reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64,
		reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64,
		reflect.Float32, reflect.Float64,
		reflect.Complex64, reflect.Complex128:
		return enc.marshallNumber(data)
	case reflect.Uintptr, reflect.UnsafePointer,
		reflect.Chan, reflect.Func:
		return []byte{}, fmt.Errorf("unable to encode kind %d", rv.Kind())
	case reflect.Invalid:
		return []byte{}, nil
	default:
		return []byte{}, fmt.Errorf("unknown kind %d", rv.Kind())
	}
}

func (enc *Encoder) marshallPtr(data interface{}) ([]byte, error) {
	// Go to dereferenced value
	dereferencedRv := reflect.ValueOf(data).Elem()
	for dereferencedRv.Kind() == reflect.Ptr {
		dereferencedRv = dereferencedRv.Elem()
	}

	// Marshall dereferenced value
	dereferencedBytes, err := enc.Marshall(dereferencedRv.Interface())
	if err != nil {
		return []byte{}, err
	}

	results := append([]byte{}, dereferencedBytes...)
	return results, nil
}

func (enc *Encoder) marshallTime(data interface{}) ([]byte, error) {
	// Format: [kind (8-bit)][UnixMilli as int]
	// Cast to time.Time
	timeData, ok := data.(time.Time)
	if !ok {
		return []byte{}, fmt.Errorf("invalid time: %s", timeData)
	}

	// Convert to int64 value (epoch)
	kindByte := byte(Time)
	timeDataAsBytes, err := enc.marshallNumber(timeData.UnixMilli())
	if err != nil {
		return []byte{}, err
	}

	results := append([]byte{kindByte}, timeDataAsBytes...)
	return results, nil
}

func (enc *Encoder) marshallIterable(data interface{}) ([]byte, error) {
	// Format: [kind (8-bit)][#bytes(8-bit) for #items][#items][items]
	dataRv := reflect.ValueOf(data)
	kindByte := byte(dataRv.Kind())
	numItems := dataRv.Len()
	numItemsBytes, err := enc.packageNum(numItems)
	if err != nil {
		return []byte{}, err
	}

	results := append([]byte{kindByte}, numItemsBytes...)
	for i := 0; i < numItems; i++ {
		item := dataRv.Index(i).Interface()
		itemBytes, err := enc.Marshall(item)
		if err != nil {
			return []byte{}, err
		}
		results = append(results, itemBytes...)
	}
	return results, nil
}

func (enc *Encoder) marshallMap(data interface{}) ([]byte, error) {
	// Format: [kind (8-bit)][#bytes(8-bit) for #kv-pairs][#kv-pairs][kv-pairs]
	dataRv := reflect.ValueOf(data)
	kindByte := byte(dataRv.Kind())
	numPairs := dataRv.Len()
	numPairsBytes, err := enc.packageNum(numPairs) // 1 <= x <= 8
	if err != nil {
		return []byte{}, err
	}
	results := append([]byte{kindByte}, numPairsBytes...)

	for _, key := range dataRv.MapKeys() {
		// Marshall key
		keyBytes, err := enc.Marshall(key.Interface())
		if err != nil {
			return []byte{}, err
		}

		// Marshall value
		var valueBytes []byte
		value := dataRv.MapIndex(key)
		// Ref: https://stackoverflow.com/questions/18306151/in-go-which-value-s-kind-is-reflect-interface
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

		results = append(results, keyBytes...)
		results = append(results, valueBytes...)
	}
	return results, nil
}

func (enc *Encoder) marshallStruct(data interface{}) ([]byte, error) {
	// Format: [kind (8-bit)][#bytes(8-bit) for #fields][#fields][field-value pairs]
	rv := reflect.ValueOf(data)
	kindByte := byte(rv.Kind())
	numFields := rv.NumField()
	numFieldsBytes, err := enc.packageNum(numFields)
	if err != nil {
		return []byte{}, err
	}

	results := append([]byte{kindByte}, numFieldsBytes...)

	structType := rv.Type()
	for i := 0; i < numFields; i++ {
		// Marshall Field Name
		fieldKey := structType.Field(i)
		keyBytes, err := enc.Marshall(fieldKey.Name)
		if err != nil {
			return []byte{}, err
		}

		// Marshall Field Value
		var valueBytes []byte
		fieldValue := rv.Field(i)
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

		results = append(results, keyBytes...)
		results = append(results, valueBytes...)
	}
	return results, nil
}

func (enc *Encoder) marshallInterface(data interface{}) ([]byte, error) {
	// Format: [kind (8-bit)][#bytes(8-bit) for length][length][value]
	/*
		- Saves its dynamic data as raw byte slice to be handled by Decoder
		- Usually only called within maps, structs or iterables
	*/
	kindByte := byte(reflect.Interface)
	dataBytes, err := enc.Marshall(data)
	if err != nil {
		return []byte{}, err
	}

	dataLenBytes, err := enc.packageNum(len(dataBytes))
	if err != nil {
		return []byte{}, err
	}

	results := append([]byte{kindByte}, dataLenBytes...)
	results = append(results, dataBytes...)
	return results, nil
}

func (enc *Encoder) marshallString(data interface{}) ([]byte, error) {
	// Format: [kind (8-bit)][#bytes(8-bit) for length][length][value]
	// Convert aliases to the underlying primitive
	data = reflect.ValueOf(data).Convert(kindToType[reflect.String]).Interface()
	kindByte := byte(reflect.String)
	dataStr, ok := data.(string)
	if !ok {
		return []byte{}, errors.New("uncastable to string")
	}

	dataBytes := []byte(dataStr)
	dataLenBytes, err := enc.packageNum(len(dataBytes))
	if err != nil {
		return []byte{}, err
	}

	results := []byte{kindByte}
	results = append(results, dataLenBytes...)
	results = append(results, dataBytes...)
	return results, nil
}

func (enc *Encoder) marshallNumber(data interface{}) ([]byte, error) {
	// Format: [kind (8-bit)][#bytes (8-bit)][value]
	// Convert aliases to the underlying primitive
	primitiveType := reflect.TypeOf(data).Kind()
	data = reflect.ValueOf(data).Convert(kindToType[primitiveType]).Interface()
	kindByte := byte(reflect.ValueOf(data).Kind())
	dataBytes, err := enc.packageNum(data)
	if err != nil {
		return []byte{}, err
	}
	results := []byte{kindByte}
	results = append(results, dataBytes...)
	return results, nil
}

func (enc *Encoder) packageNum(data interface{}) ([]byte, error) {
	// Returns [#bytes (8-bit)][value]
	/*
		- If data is a float, marshall directly to bytes without doing anything
			- Decoder should decode according to dest type

		- If data's value is negative, dest MUST interpret the bytes using an exact same type,
		otherwise may interpret it as a positive value
			- Instead of `compress` parameter, can alternatively check if data is negative

		- If data value is positive, can cast it to be as small as possible
			- dest should use a large enough data value to make sure interpreted
			value is correct.
	*/

	var castedData interface{}
	kind := reflect.TypeOf(data).Kind() // In case data is an alias
	switch kind {
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		dataAsInt := int(reflect.ValueOf(data).Convert(kindToType[reflect.Int]).Int())
		if dataAsInt >= 0 {
			castedData = enc.getSmallestIntType(dataAsInt)
		} else if kind == reflect.Int {
			numBytes := reflect.TypeOf(int(0)).Bits() / 8
			if numBytes == 1 {
				castedData = int8(dataAsInt)
			} else if numBytes == 2 {
				castedData = int16(dataAsInt)
			} else if numBytes == 4 {
				castedData = int32(dataAsInt)
			} else {
				castedData = int64(dataAsInt)
			}
		} else {
			castedData = data
		}
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		dataAsUint := uint(reflect.ValueOf(data).Convert(kindToType[reflect.Uint]).Uint())
		castedData = enc.getSmallestUintType(dataAsUint)
	case reflect.Float32, reflect.Float64:
		castedData = data
	default:
		return []byte{}, fmt.Errorf("kind %s is not a number", kind)
	}

	buf := bytes.NewBuffer([]byte{})
	err := binary.Write(buf, endianness, castedData)
	if err != nil {
		return []byte{}, err
	}

	numBytesLen := byte(buf.Len())
	results := []byte{numBytesLen}
	dataBytes := buf.Bytes()
	return append(results, dataBytes...), nil
}

func (enc *Encoder) getSmallestIntType(num int) interface{} {
	if num >= -128 && num <= 127 {
		return int8(num)
	} else if num >= -32768 && num <= 32767 {
		return int16(num)
	} else if num >= -2147483648 && num <= 2147483647 {
		return int32(num)
	} else {
		// Range: -9223372036854775808 through 9223372036854775807.
		return int64(num)
	}
}

func (enc *Encoder) getSmallestUintType(num uint) interface{} {
	if num <= 255 {
		return uint8(num)
	} else if num <= 65535 {
		return uint16(num)
	} else if num <= 4294967295 {
		return uint32(num)
	} else {
		// Range: 0 through 18446744073709551615.
		return uint64(num)
	}
}
