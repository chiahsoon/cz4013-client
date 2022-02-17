package services

import (
	"fmt"
	"reflect"
	"strings"
)

const printBoundary = "================================="

type PrettyPrinter struct{}

func (pp *PrettyPrinter) Print(data interface{}, header, footer string) {
	rv := reflect.ValueOf(data)
	for rv.Kind() == reflect.Ptr || rv.Kind() == reflect.Interface {
		rv = rv.Elem()
	}

	if !rv.IsValid() {
		return
	}

	fn := func(interface{}) {
		if rv.Kind() == reflect.Struct {
			pp.printStruct(rv.Interface())
		} else if rv.Kind() == reflect.Map {
			pp.printMap(rv.Interface())
		} else {
			fmt.Println(data)
		}
	}
	pp.printBoundary(fn, header, footer, data)
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

func (pp *PrettyPrinter) printBoundary(printDataFn func(interface{}), header, footer string, data interface{}) {
	fmt.Println(printBoundary)
	if len(header) > 0 {
		fmt.Println(header)
	}
	printDataFn(data)
	if len(footer) > 0 {
		fmt.Println(footer)
	}
	fmt.Println(printBoundary)
}

func (pp *PrettyPrinter) printStruct(data interface{}) {
	s := reflect.ValueOf(data)
	dataType := s.Type()

	for i := 0; i < s.NumField(); i++ {
		field := s.Field(i)
		line := fmt.Sprintf("%s:  %v", dataType.Field(i).Name, field.Interface())
		fmt.Println(strings.TrimSuffix(line, "\n"))
	}
}

func (pp *PrettyPrinter) printMap(data interface{}) {
	if dataMap, ok := data.(map[string]interface{}); ok {
		for key, value := range dataMap {
			line := fmt.Sprintf("%s:  %v", key, value)
			fmt.Println(strings.TrimSuffix(line, "\n"))
		}
		return
	}
	fmt.Println(data)
}
