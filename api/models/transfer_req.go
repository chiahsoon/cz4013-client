package models

type TransferReq struct {
	AccountNumber     int
	Name              string
	Password          string
	Currency          string
	Amount            float64
	DestAccountNumber int
}
