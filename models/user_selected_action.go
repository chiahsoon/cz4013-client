package models

import "errors"

type UserSelectedAction int

const (
	OpenAccountAction UserSelectedAction = iota
	CloseAccountAction
	GetBalanceAction
	DepositAction
	WithdrawAction
	MonitorAction
)

var AllActions = []UserSelectedAction{
	OpenAccountAction,
	CloseAccountAction,
	GetBalanceAction,
	DepositAction,
	WithdrawAction,
	MonitorAction,
}

func (a UserSelectedAction) IsValid() error {
	for _, validAction := range AllActions {
		if validAction == a {
			return nil
		}
	}

	return errors.New("invalid action")
}

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
	default:
		return "Unknown action"
	}
}
