package codec

import (
	"bytes"
	"encoding/binary"
	"errors"
	"fmt"
	"reflect"
	"time"
)

var (
	endianness = binary.BigEndian
)

type Decoder struct{}

func (dec *Decoder) UnmarshallFromInterface(src interface{}, dest interface{}) error {
	/*
		- When encoded by Encoder, interface fields are left as bytes
		- When Decoder decodes structs/maps with these fields, they are left as bytes (see unmarshallInterface)
	*/
	if reflect.ValueOf(dest).Kind() != reflect.Ptr {
		return errors.New("dest is not a ptr")
	}

	bytes, ok := src.([]byte)
	if !ok {
		return errors.New("src interface not encoded as bytes")
	}

	return dec.Unmarshall(bytes, dest)
}

func (dec *Decoder) Unmarshall(data []byte, dest interface{}) error {
	if reflect.ValueOf(dest).Kind() != reflect.Ptr {
		return errors.New("dest is not a ptr")
	}
	buf := bytes.NewBuffer(data)
	return dec.unmarshall(buf, dest)
}

func (dec *Decoder) unmarshall(buf *bytes.Buffer, dest interface{}) error {
	kindVal, err := buf.ReadByte()
	if err != nil {
		return err
	}

	kind := reflect.Kind(kindVal)
	switch kind {
	case reflect.Ptr:
		return dec.unmarshallPtr(buf, dest)
	case reflect.Interface:
		return dec.unmarshallInterface(buf, dest)
	case Time:
		return dec.unmarshallTime(buf, dest)
	case reflect.Struct:
		return dec.unmarshallStruct(buf, dest)
	case reflect.Map:
		return dec.unmarshallMap(buf, dest)
	case reflect.Array, reflect.Slice:
		return dec.unmarshallIterable(buf, dest)
	case reflect.String:
		return dec.unmarshallString(buf, dest)
	case reflect.Bool,
		reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64,
		reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64,
		reflect.Float32, reflect.Float64,
		reflect.Complex64, reflect.Complex128:
		return dec.unmarshallNumber(buf, dest)
	case reflect.Invalid,
		reflect.Uintptr, reflect.UnsafePointer,
		reflect.Chan, reflect.Func:
		return fmt.Errorf("unable to decode kind %d", kind)
	default:
		return fmt.Errorf("unknown kind %d", kind)
	}
}

func (dec *Decoder) unmarshallPtr(buf *bytes.Buffer, dest interface{}) error {
	// dest is at least **type
	destRv := reflect.ValueOf(dest).Elem()

	dereferencedType := reflect.TypeOf(dest).Elem().Elem()
	destRv.Set(reflect.New(dereferencedType)) // Initialise address
	return dec.unmarshall(buf, destRv.Interface())
}

func (dec *Decoder) unmarshallTime(buf *bytes.Buffer, dest interface{}) error {
	// Format: [kind (8-bit)][UnixMilli as int]
	var timeUnixData int64
	if err := dec.unmarshall(buf, &timeUnixData); err != nil {
		return err
	}

	timeData := time.UnixMilli(timeUnixData)
	timePtrRv := reflect.ValueOf(dest)
	timeRv := timePtrRv.Elem()

	// Go to dereferenced value and initialise
	destType := reflect.TypeOf(dest).Elem()
	if destType.Kind() == reflect.Ptr {
		for destType.Kind() == reflect.Ptr {
			destType = destType.Elem()
		}
		timePtrRv.Elem().Set(reflect.New(destType))

		timeRv = timePtrRv.Elem()
		for timeRv.Kind() == reflect.Ptr {
			timeRv = timeRv.Elem()
		}
	}

	timeRv.Set(reflect.ValueOf(timeData))
	return nil
}

func (dec *Decoder) unmarshallInterface(buf *bytes.Buffer, dest interface{}) error {
	// Format: [kind (8-bit)][#bytes(8-bit) for length][length][value]
	dataLength, err := dec.readLength64(buf)
	if err != nil {
		return err
	}

	data := make([]byte, dataLength)
	_, err = buf.Read(data)
	if err != nil {
		return err
	}

	destRv := reflect.ValueOf(dest).Elem()
	destRv.Set(reflect.ValueOf(data))
	return nil
}

func (dec *Decoder) unmarshallStruct(buf *bytes.Buffer, dest interface{}) error {
	// Format: [kind (8-bit)][#bytes(8-bit) for #fields][#fields][field-value pairs]
	numFields, err := dec.readLength64(buf)
	if err != nil {
		return err
	}

	structPtrRv := reflect.ValueOf(dest)
	structRv := structPtrRv.Elem()

	// Go to dereferenced value and initialise
	destType := reflect.TypeOf(dest).Elem()
	if destType.Kind() == reflect.Ptr {
		for destType.Kind() == reflect.Ptr {
			destType = destType.Elem()
		}
		structPtrRv.Elem().Set(reflect.New(destType))

		structRv = structPtrRv.Elem()
		for structRv.Kind() == reflect.Ptr {
			structRv = structRv.Elem()
		}
	}

	for idx := 0; idx < int(numFields); idx++ {
		// Unmarshall field name
		var fieldName string
		if err := dec.unmarshall(buf, &fieldName); err != nil {
			return err
		}

		// Initialise and unmarshall field value
		structFieldRv := structRv.FieldByName(fieldName)
		structFieldType := structFieldRv.Type()
		valuePtrRv := reflect.New(structFieldType)

		valuePtr := valuePtrRv.Interface()
		if err := dec.unmarshall(buf, valuePtr); err != nil {
			return err
		}

		valueRv := valuePtrRv.Elem()
		structFieldRv.Set(valueRv)
	}

	return nil
}

func (dec *Decoder) unmarshallMap(buf *bytes.Buffer, dest interface{}) error {
	// Format: [kind (8-bit)][#bytes(8-bit) for #kv-pairs][#kv-pairs][kv-pairs]
	/*
		- If dest key is interface{} type, it will be left as bytes
		- Therefore, impossible to index with the actual value of the key because
		of extra encoding data (kind, length, etc.)
	*/

	numPairs, err := dec.readLength64(buf)
	if err != nil {
		return err
	}

	mapPtrRv := reflect.ValueOf(dest)
	mapRv := mapPtrRv.Elem()

	// Go to dereferenced value and initialise
	destType := reflect.TypeOf(dest).Elem()
	if destType.Kind() == reflect.Ptr {
		for destType.Kind() == reflect.Ptr {
			destType = destType.Elem()
		}
		mapPtrRv.Elem().Set(reflect.New(destType))

		mapRv = mapPtrRv.Elem()
		for mapRv.Kind() == reflect.Ptr {
			mapRv = mapRv.Elem()
		}
	}

	mapType := mapRv.Type()
	keyType := mapType.Key()
	valueType := mapType.Elem()
	// Initialise map to avoid nil errors
	nonNilMap := reflect.MakeMap(mapType)
	mapRv.Set(nonNilMap)

	for idx := 0; idx < int(numPairs); idx++ {
		// Unmarshall key
		keyPtrRv := reflect.New(keyType)
		keyPtr := keyPtrRv.Interface()
		if err := dec.unmarshall(buf, keyPtr); err != nil {
			return err
		}

		// Unmarshall value
		valuePtrRv := reflect.New(valueType)
		valuePtr := valuePtrRv.Interface()
		if err := dec.unmarshall(buf, valuePtr); err != nil {
			return err
		}

		keyRv := keyPtrRv.Elem()
		valueRv := valuePtrRv.Elem()
		mapRv.SetMapIndex(keyRv, valueRv)
	}

	return nil
}

func (dec *Decoder) unmarshallIterable(buf *bytes.Buffer, dest interface{}) error {
	// Format: [kind (8-bit)][#bytes(8-bit) for length][length][value]
	numItems, err := dec.readLength64(buf)
	if err != nil {
		return err
	}

	iterablePtrRv := reflect.ValueOf(dest)
	iterableRv := iterablePtrRv.Elem()
	itemType := iterableRv.Type()

	// Go to dereferenced value and initialise
	destType := reflect.TypeOf(dest).Elem()
	if destType.Kind() == reflect.Ptr {
		for destType.Kind() == reflect.Ptr {
			destType = destType.Elem()
		}
		iterablePtrRv.Elem().Set(reflect.New(destType))

		iterableRv = iterablePtrRv.Elem()
		for iterableRv.Kind() == reflect.Ptr {
			iterableRv = iterableRv.Elem()
		}

		itemType = iterableRv.Type()
	}

	// Initialise to avoid out-of-bounds error
	size := int(numItems)
	newSizedSlice := reflect.MakeSlice(itemType, size, size)
	iterableRv.Set(newSizedSlice)

	for idx := 0; idx < int(numItems); idx++ {
		// Unmarshall item
		itemRv := iterableRv.Index(idx)
		itemPtrRv := itemRv.Addr()
		itemPtr := itemPtrRv.Interface()
		if err = dec.unmarshall(buf, itemPtr); err != nil {
			return err
		}
	}

	return nil
}

func (dec *Decoder) unmarshallString(buf *bytes.Buffer, dest interface{}) error {
	// Format: [kind (8-bit)][#bytes(8-bit) for length][length][value]
	length, err := dec.readLength64(buf)
	if err != nil {
		return err
	}

	strDataBytes := make([]byte, length)
	_, err = buf.Read(strDataBytes)
	if err != nil {
		return err
	}

	strPtrRv := reflect.ValueOf(dest)
	strRv := strPtrRv.Elem()

	// Go to dereferenced value and initialise
	destType := reflect.TypeOf(dest).Elem()
	if destType.Kind() == reflect.Ptr {
		for destType.Kind() == reflect.Ptr {
			destType = destType.Elem()
		}
		strPtrRv.Elem().Set(reflect.New(destType))

		strRv = strPtrRv.Elem()
		for strRv.Kind() == reflect.Ptr {
			strRv = strRv.Elem()
		}
	}

	// Dealing with aliases
	strRv.Set(reflect.ValueOf(string(strDataBytes)).Convert(strRv.Type()))
	return nil
}

func (dec *Decoder) unmarshallNumber(buf *bytes.Buffer, dest interface{}) error {
	// Format: [kind (8-bit)][#bytes (8-bit)][value]
	numByteForNumVal, err := buf.ReadByte()
	if err != nil {
		return err
	}

	data := make([]byte, numByteForNumVal)
	_, err = buf.Read(data)
	if err != nil {
		return err
	}

	numPtrRv := reflect.ValueOf(dest)
	destType := reflect.TypeOf(dest).Elem()
	numRv := numPtrRv.Elem()

	// Go to dereferenced value and initialise
	if destType.Kind() == reflect.Ptr {
		for destType.Kind() == reflect.Ptr {
			destType = destType.Elem()
		}

		numPtrRv.Elem().Set(reflect.New(destType))

		numRv = numPtrRv.Elem()
		for numRv.Kind() == reflect.Ptr {
			numRv = numRv.Elem()
		}
	}

	// Deal with different type in encoded data and dest type
	var finalVal interface{}
	if numRv.Kind() == reflect.Float32 || numRv.Kind() == reflect.Float64 {
		if destType.Kind() == reflect.Float64 && numByteForNumVal == 4 {
			numVal, err := dec.bytesToSizedNum(data, reflect.TypeOf(float32(0)))
			if err != nil {
				return err
			}
			finalVal = float64(numVal.(float32))
		} else if destType.Kind() == reflect.Float32 && numByteForNumVal == 8 {
			numVal, err := dec.bytesToSizedNum(data, reflect.TypeOf(float64(0)))
			if err != nil {
				return err
			}
			finalVal = float32(numVal.(float64))
		} else {
			finalVal, err = dec.bytesToSizedNum(data, destType)
			if err != nil {
				return err
			}
		}
	} else {
		// Reduce/Pad integer values
		destSize := destType.Bits() / 8
		if destSize < int(numByteForNumVal) {
			data = data[len(data)-destSize:]
		} else if destSize > int(numByteForNumVal) {
			// Pad with 0 (pos) or 255 (neg)
			pad := make([]byte, destSize-len(data))
			if data[0] >= 128 {
				for idx := 0; idx < len(pad); idx++ {
					pad[idx] = 255
				}
			}
			data = append(pad, data...)
		}

		numVal, err := dec.bytesToSizedNum(data, destType)
		if err != nil {
			return err
		}

		// Cast to int/uint (if necessary) - machine dependent types
		finalVal = numVal
		if destType.Kind() == reflect.Int {
			finalVal = reflect.ValueOf(numVal).Convert(kindToType[reflect.Int]).Interface()
		} else if destType.Kind() == reflect.Uint {
			finalVal = reflect.ValueOf(numVal).Convert(kindToType[reflect.Uint]).Interface()
		}
	}

	// Dealing with aliases
	numRv.Set(reflect.ValueOf(finalVal).Convert(numRv.Type()))
	return nil
}

func (dec *Decoder) readLength64(buf *bytes.Buffer) (int64, error) {
	numBytesForLength, err := buf.ReadByte()
	if err != nil {
		return 0, nil
	}

	bytesForLength := make([]byte, numBytesForLength)
	_, err = buf.Read(bytesForLength)
	if err != nil {
		return 0, err
	}

	numAsInterface, err := dec.bytesToSizedNum(bytesForLength, kindToType[reflect.Int64])
	if err != nil {
		return 0, err
	}

	num, ok := numAsInterface.(int64)
	if !ok {
		return 0, errors.New("failed to convert length interface to int64")
	}
	return num, nil
}

func (dec *Decoder) bytesToSizedNum(data []byte, targetType reflect.Type) (interface{}, error) {
	if targetType.Kind() == reflect.Ptr || targetType.Kind() == reflect.Interface {
		return nil, errors.New("cannot convert bytes to dest type pointer or interface")
	}

	// If target is not fixed, change to fixed type of same nature
	targetNumBytes := targetType.Bits() / 8
	resultsPtrRv := reflect.New(targetType)
	if targetType.Kind() == reflect.Int {
		resultsPtrRv = dec.getIntRvForNumBytes(targetNumBytes)
	} else if targetType.Kind() == reflect.Uint {
		resultsPtrRv = dec.getUintRvForNumBytes(targetNumBytes)
	}

	// Pad with 0s if targetType requires more bytes
	resultsData := make([]byte, 0)
	if endianness == binary.BigEndian {
		numBytesRequried := targetNumBytes
		for i := 0; i < numBytesRequried-len(data); i++ {
			resultsData = append(resultsData, 0)
		}

		resultsData = append(resultsData, data...)
	} else {
		resultsData = append(resultsData, data...)
		numBytesRequried := targetNumBytes
		for i := 0; i < numBytesRequried-len(data); i++ {
			resultsData = append(resultsData, 0)
		}
	}

	buf := bytes.NewBuffer(resultsData)
	if err := binary.Read(buf, endianness, resultsPtrRv.Interface()); err != nil {
		return nil, err
	}

	return resultsPtrRv.Elem().Interface(), nil
}

func (dec *Decoder) getIntRvForNumBytes(numBits int) reflect.Value {
	if numBits == 1 {
		return reflect.ValueOf(new(int8))
	} else if numBits == 2 {
		return reflect.ValueOf(new(int16))
	} else if numBits == 4 {
		return reflect.ValueOf(new(int32))
	}
	return reflect.ValueOf(new(int64))
}

func (dec *Decoder) getUintRvForNumBytes(numBits int) reflect.Value {
	if numBits == 1 {
		return reflect.ValueOf(new(uint8))
	} else if numBits == 2 {
		return reflect.ValueOf(new(uint16))
	} else if numBits == 4 {
		return reflect.ValueOf(new(uint32))
	}
	return reflect.ValueOf(new(uint64))
}
