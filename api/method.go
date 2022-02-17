package api

import "errors"

type APIMethod string

const (
	OpenAccountAPI   APIMethod = "open"
	CloseAccountAPI  APIMethod = "close"
	GetBalanceAPI    APIMethod = "balance"
	UpdateBalanceAPI APIMethod = "update_balance"
	MonitorAPI       APIMethod = "monitor"
	CheckStateAPI    APIMethod = "check_state"
)

func (m APIMethod) Validate() error {
	switch m {
	case OpenAccountAPI, CloseAccountAPI, GetBalanceAPI, UpdateBalanceAPI, MonitorAPI, CheckStateAPI:
		return nil
	}
	return errors.New("invalid api method")
}
