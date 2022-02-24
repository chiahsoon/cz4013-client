package models

import "fmt"

type Account struct {
	Number     int
	HolderName string
	Password   string
	Currency   Currency
	Balance    float64
}

func (acc *Account) GetDetails() string {
	details := ""
	details += fmt.Sprintf("Account Number: %d\n", acc.Number)
	details += fmt.Sprintf("Account Holder Name: %s\n", acc.HolderName)
	details += fmt.Sprintf("Account Currency: %s\n", acc.Currency)
	details += fmt.Sprintf("Account Balance: %f\n", acc.Balance)
	return details
}
