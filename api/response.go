package api

import (
	"fmt"
	"strings"
)

type Response struct {
	ErrMsg string
	Data   interface{}
}

func (r *Response) HasError() bool {
	return r.ErrMsg != ""
}

func (r *Response) Display() {
	if r.HasError() {
		fmt.Println(strings.Title(r.ErrMsg) + "\n")
		return
	}
	fmt.Println(r.Data)
	fmt.Println()
}
