package services

import (
	"fmt"
	"reflect"
	"strings"
)

const printBoundary = "================================="

type PrettyPrinter struct{}

func (pp *PrettyPrinter) Print(data interface{}, header, footer string) {
	pp.printBoundary(pp.printByKind, header, footer, data)
}

func (pp *PrettyPrinter) PrintMessage(msg, header, footer string) {
	fn := func(interface{}) {
		fmt.Println(strings.Title(msg))
	}
	pp.printBoundary(fn, header, footer, msg)
}

func (pp *PrettyPrinter) PrintError(errMsg, header, footer string) {
	pp.PrintMessage("Error: "+errMsg, header, footer)
}

func (pp *PrettyPrinter) printByKind(data interface{}) {
	rv := reflect.ValueOf(data)
	switch rv.Kind() {
	case reflect.Ptr, reflect.Interface:
		for rv.Kind() == reflect.Ptr || rv.Kind() == reflect.Interface {
			rv = rv.Elem()
		}
		pp.printByKind(rv.Interface())
	case reflect.Struct:
		pp.printStruct(rv.Interface())
	case reflect.Map:
		pp.printMap(rv.Interface())
	case reflect.Slice:
		pp.printSlice(rv.Interface())
	default:
		fmt.Println(data)
	}
}

func (pp *PrettyPrinter) printStruct(data interface{}) {
	// TODO: Assumes values are primitives/structs
	s := reflect.ValueOf(data)
	dataType := s.Type()

	for i := 0; i < s.NumField(); i++ {
		field := s.Field(i)
		line := fmt.Sprintf("%s:  %v", dataType.Field(i).Name, field.Interface())
		fmt.Println(strings.TrimSuffix(line, "\n"))
	}
}

func (pp *PrettyPrinter) printMap(data interface{}) {
	// TODO: Assumes values are primitives/structs
	if dataMap, ok := data.(map[string]interface{}); ok {
		for key, value := range dataMap {
			line := fmt.Sprintf("%s:  %v", key, value)
			fmt.Println(strings.TrimSuffix(line, "\n"))
		}
		return
	}
	fmt.Println(data)
}

func (pp *PrettyPrinter) printSlice(data interface{}) {
	rv := reflect.ValueOf(data)
	for i := 0; i < rv.Len(); i++ {
		sliceItem := rv.Index(i).Interface()
		sliceItemRv := reflect.ValueOf(sliceItem)
		if i > 0 {
			fmt.Println()
		}

		pp.printByKind(sliceItemRv.Interface())
	}
}

func (pp *PrettyPrinter) printBoundary(printDataFn func(interface{}), header, footer string, data interface{}) {
	fmt.Println(printBoundary)
	if len(header) > 0 {
		fmt.Println(header)
	}
	printDataFn(data)
	if len(footer) > 0 {
		fmt.Println(footer)
	}
	// fmt.Println(printBoundary)
}
