package models

type UpdateBalanceReq struct {
	AccountNumber int
	Name          string
	Password      string
	Currency      string
	Amount        float64
}
