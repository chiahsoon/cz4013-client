package models

type OpenAccountReq struct {
	AccountNumber  int
	Name           string
	Password       string
	Currency       string
	InitialBalance float64
}
