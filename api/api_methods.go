package api

import "errors"

type APIMethod string

const (
	OpenAccountAPI   APIMethod = "open"
	CloseAccountAPI  APIMethod = "close"
	GetBalanceAPI    APIMethod = "balance"
	UpdateBalanceAPI APIMethod = "update_balance"
	MonitorMethodAPI APIMethod = "monitor"
)

func (m APIMethod) IsValid() error {
	switch m {
	case OpenAccountAPI, CloseAccountAPI, UpdateBalanceAPI, MonitorMethodAPI:
		return nil
	}
	return errors.New("invalid api method")
}
