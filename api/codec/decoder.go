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

// !Does not allow map[interface{}]interface{}
type Decoder struct{}

func (dec *Decoder) UnmarshallFromInterface(src interface{}, dest interface{}) error {
	// If an interface is encoded by Encoder, interface fields will be encoded with bytes underneath
	// When Decoder decodes structs/maps with these fields, they are left as bytes (see unmarshallInterface)

	if reflect.ValueOf(dest).Kind() != reflect.Ptr {
		return errors.New("dest is not a ptr")
	}

	bytes, ok := src.([]byte)
	if !ok {
		return errors.New("src is not a byte slice nor string")
	}

	return dec.Unmarshall(bytes, dest)
}

func (dec *Decoder) Unmarshall(data []byte, dest interface{}) error {
	// Struct fields or map values cannot be interface{}
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
	case 27:
		return dec.unmarshallTime(buf, dest)
	case reflect.Invalid,
		reflect.Uintptr, reflect.UnsafePointer,
		reflect.Chan, reflect.Func:
		return fmt.Errorf("unable to decode kind %d", kind)
	default:
		return fmt.Errorf("unknown kind %d", kind)
	}
}

func (dec *Decoder) unmarshallPtr(buf *bytes.Buffer, dest interface{}) error {
	// dest will be at least **type
	rv := reflect.ValueOf(dest)
	ptrRv := rv.Elem()
	dereferencedType := reflect.TypeOf(dest).Elem().Elem()
	ptrRv.Set(reflect.New(dereferencedType))
	return dec.unmarshall(buf, ptrRv.Interface())
}

func (dec *Decoder) unmarshallTime(buf *bytes.Buffer, dest interface{}) error {
	var timeUnixData int64
	if err := dec.unmarshall(buf, &timeUnixData); err != nil {
		return err
	}

	timeData := time.UnixMilli(timeUnixData)
	timePtrRv := reflect.ValueOf(dest)
	timeRv := timePtrRv.Elem()
	// !CHANGE
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
	// !END CHANGE

	timeRv.Set(reflect.ValueOf(timeData))
	return nil
}

func (dec *Decoder) unmarshallInterface(buf *bytes.Buffer, dest interface{}) error {
	dataLength, err := dec.readInt64(buf)
	if err != nil {
		return err
	}

	data := make([]byte, dataLength)
	_, err = buf.Read(data)
	if err != nil {
		return err
	}

	// TODO: Follow up
	bytesPtrRv := reflect.ValueOf(dest)
	bytesRv := bytesPtrRv.Elem()
	bytesRv.Set(reflect.ValueOf(data))
	return nil
}

func (dec *Decoder) unmarshallStruct(buf *bytes.Buffer, dest interface{}) error {
	numFields, err := dec.readInt8(buf)
	if err != nil {
		return err
	}

	structPtrRv := reflect.ValueOf(dest)
	structRv := structPtrRv.Elem()

	// !CHANGE
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
	// !END CHANGE

	for idx := 0; idx < int(numFields); idx++ {
		var fieldName string
		if err := dec.unmarshall(buf, &fieldName); err != nil {
			return err
		}
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
	// Format: [number (64-bits) of pairs][pairs]
	numPairs, err := dec.readInt64(buf)
	if err != nil {
		return err
	}

	mapPtrRv := reflect.ValueOf(dest)
	mapRv := mapPtrRv.Elem()
	var mapType, keyType, valueType reflect.Type

	// !CHANGE
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

		mapType = mapRv.Type()
		keyType = mapType.Key()
		valueType = mapType.Elem()
	} else {
		mapType = mapRv.Type()
		// If keyType is interface{}, impossible to get desired type
		keyType = mapType.Key()
		valueType = mapType.Elem()
	}
	// !END CHANGE

	// Avoid nil errors
	nonNilMap := reflect.MakeMap(mapType)
	mapRv.Set(nonNilMap)

	for idx := 0; idx < int(numPairs); idx++ {
		keyPtrRv := reflect.New(keyType)
		keyPtr := keyPtrRv.Interface()
		if err := dec.unmarshall(buf, keyPtr); err != nil {
			return err
		}

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
	// Format: [number (64-bits) of items][iterable items]
	numItems, err := dec.readInt64(buf)
	if err != nil {
		return err
	}

	iterablePtrRv := reflect.ValueOf(dest)
	iterableRv := iterablePtrRv.Elem()
	itemType := iterableRv.Type()

	// !CHANGE
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
	// !END CHANGE

	// Avoid out-of-bounds error
	newSizedSlice := reflect.MakeSlice(itemType, int(numItems), int(numItems))
	iterableRv.Set(newSizedSlice)

	for idx := 0; idx < int(numItems); idx++ {
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
	// Format: [value length (64-bit)][value]
	length, err := dec.readInt64(buf)
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

	// !CHANGE
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
	// !END CHANGE

	// Dealing with aliases
	strRv.Set(reflect.ValueOf(string(strDataBytes)).Convert(strRv.Type()))
	return nil
}

func (dec *Decoder) unmarshallNumber(buf *bytes.Buffer, dest interface{}) error {
	// Format: [number length (8-bit)][value]

	length, err := dec.readInt8(buf)
	if err != nil {
		return err
	}
	data := make([]byte, length)
	_, err = buf.Read(data)
	if err != nil {
		return err
	}

	numPtrRv := reflect.ValueOf(dest)
	destType := reflect.TypeOf(dest).Elem()
	numRv := numPtrRv.Elem()

	// !CHANGE
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
	// !END CHANGE

	numVal, err := dec.bytesToFixedNum(data, destType)
	if err != nil {
		return err
	}

	// Cast to int/uint (if necessary) which are machine dependent types
	var finalVal interface{}
	if destType.Kind() == reflect.Int {
		switch reflect.ValueOf(numVal).Kind() {
		case reflect.Int8:
			finalVal = int(numVal.(int8))
		case reflect.Int16:
			finalVal = int(numVal.(int16))
		case reflect.Int32:
			finalVal = int(numVal.(int32))
		case reflect.Int64:
			finalVal = int(numVal.(int64))
		default:
			return fmt.Errorf("invalid int type: %d", reflect.ValueOf(numVal).Kind())
		}
	} else if destType.Kind() == reflect.Uint {
		switch reflect.ValueOf(numVal).Kind() {
		case reflect.Uint8:
			finalVal = uint(numVal.(uint8))
		case reflect.Uint16:
			finalVal = uint(numVal.(uint16))
		case reflect.Uint32:
			finalVal = uint(numVal.(uint32))
		case reflect.Uint64:
			finalVal = uint(numVal.(uint64))
		default:
			return fmt.Errorf("invalid uint type: %d", reflect.ValueOf(numVal).Kind())
		}
	} else {
		finalVal = numVal
	}

	// Dealing with aliases
	numRv.Set(reflect.ValueOf(finalVal).Convert(numRv.Type()))
	return nil
}

func (dec *Decoder) readInt64(buf *bytes.Buffer) (int64, error) {
	dataBytes := make([]byte, 8)
	_, err := buf.Read(dataBytes)
	if err != nil {
		return 0, err
	}

	var num int64
	numI, err := dec.bytesToFixedNum(dataBytes, reflect.TypeOf(num))
	if err != nil {
		return 0, err
	}

	num, ok := numI.(int64)
	if !ok {
		return 0, errors.New("failed to convert length interface to int64")
	}
	return num, nil
}

func (dec *Decoder) readInt8(buf *bytes.Buffer) (int8, error) {
	lengthByte, err := buf.ReadByte()
	if err != nil {
		return 0, err
	}

	return int8(lengthByte), nil
}

func (dec *Decoder) bytesToFixedNum(data []byte, targetType reflect.Type) (interface{}, error) {
	if targetType.Kind() == reflect.Ptr || targetType.Kind() == reflect.Interface {
		return nil, errors.New("cannot convert bytes to dest type pointer or interface")
	}

	// If target is not fixed, change to fixed type of same nature
	// TODO Confirm fixed num size
	resPtrRv := reflect.New(targetType)
	if targetType.Kind() == reflect.Int {
		resPtrRv = reflect.ValueOf(new(int64))
	} else if targetType.Kind() == reflect.Uint {
		resPtrRv = reflect.ValueOf(new(uint64))
	}

	buf := bytes.NewBuffer(data)
	resPtr := resPtrRv.Interface()
	if err := binary.Read(buf, endianness, resPtr); err != nil {
		return nil, err
	}

	return resPtrRv.Elem().Interface(), nil
}
