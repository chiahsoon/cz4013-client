package models

import "errors"

type UserSelectedAction int

// Note: Action value must match index in AllActions

const (
	OpenAccountAction UserSelectedAction = iota
	CloseAccountAction
	GetBalanceAction
	DepositAction
	WithdrawAction
	TransferAction
	MonitorAction
	CheckStateAction
)

var AllActions = []UserSelectedAction{
	OpenAccountAction,
	CloseAccountAction,
	GetBalanceAction,
	DepositAction,
	WithdrawAction,
	TransferAction,
	MonitorAction,
	CheckStateAction,
}

func (a UserSelectedAction) IsValid() error {
	for _, validAction := range AllActions {
		if validAction == a {
			return nil
		}
	}

	return errors.New("invalid action")
}

// To be used when displaying the main menu
func (a UserSelectedAction) Description() string {
	switch a {
	case OpenAccountAction:
		return "Open Account"
	case CloseAccountAction:
		return "Close Account"
	case GetBalanceAction:
		return "Retrieve Balance"
	case DepositAction:
		return "Deposit"
	case WithdrawAction:
		return "Withdraw"
	case MonitorAction:
		return "Monitor Updates"
	case CheckStateAction:
		return "Check Bank State (admin)"
	case TransferAction:
		return "Transfer Funds"
	default:
		return "Unknown action"
	}
}
