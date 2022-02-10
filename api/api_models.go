package api

import (
	"fmt"
	"strings"
)

type Request struct {
	Method APIMethod
	Data   interface{}
}

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
}

type OpenAccountReq struct {
	AccountNumber int
	Name          string
	Password      string
	Currency      string
}

type CloseAccountReq struct {
	AccountNumber int
	Name          string
	Password      string
}

type GetBalanceReq struct {
	AccountNumber int
	Name          string
	Password      string
	Currency      string
}

type UpdateBalanceReq struct {
	AccountNumber int
	Name          string
	Password      string
	Currency      string
	Amount        float64
}

type MonitorReq struct {
	Interval int
}
