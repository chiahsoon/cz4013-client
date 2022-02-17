package services

import (
	"errors"
	"strconv"

	"github.com/AlecAivazis/survey/v2"
	"github.com/chiahsoon/cz4013-client/models"
)

type UIService struct{}

func (ui *UIService) GetMainPrompt() *survey.Select {
	options := []string{}
	for i := 0; i < len(models.AllActions); i++ {
		options = append(options, models.AllActions[i].Description())
	}

	return &survey.Select{
		Message: "What would you like to do?",
		Options: append(options, "Exit"),
	}
}

func (ui *UIService) GetSubPromptsForAction() map[models.UserSelectedAction][]*survey.Question {
	return map[models.UserSelectedAction][]*survey.Question{
		models.OpenAccountAction:  {ui.getNameQn(), ui.getCurrencyQn(), ui.getPasswordQn()},
		models.CloseAccountAction: {ui.getAccountNumberQn(), ui.getNameQn(), ui.getPasswordQn()},
		models.GetBalanceAction:   {ui.getAccountNumberQn(), ui.getNameQn(), ui.getCurrencyQn(), ui.getPasswordQn()},
		models.DepositAction:      {ui.getAccountNumberQn(), ui.getNameQn(), ui.getCurrencyQn(), ui.getAmountQn(), ui.getPasswordQn()},
		models.WithdrawAction:     {ui.getAccountNumberQn(), ui.getNameQn(), ui.getCurrencyQn(), ui.getAmountQn(), ui.getPasswordQn()},
		models.MonitorAction:      {ui.getIntervalQn()},
	}
}

func (ui *UIService) getPasswordQn() *survey.Question {
	return ui.makePasswordQuestion("password", "What is your password?")
}

func (ui *UIService) getNameQn() *survey.Question {
	return ui.makeTextQuestion("name", "What is your name?")
}

func (ui *UIService) getAccountNumberQn() *survey.Question {
	return ui.makeTextQuestion("accountNumber", "What is your account number?")
}

func (ui *UIService) getCurrencyQn() *survey.Question {
	return ui.makeTextQuestion("currency", "What is the currency?")
}

func (ui *UIService) getAmountQn() *survey.Question {
	return ui.makeTextQuestion("amount", "What is the amount?")
}

func (ui *UIService) makeTextQuestion(name string, message string) *survey.Question {
	return &survey.Question{
		Name:     name,
		Prompt:   &survey.Input{Message: message},
		Validate: survey.Required,
	}
}

func (ui *UIService) makePasswordQuestion(name string, message string) *survey.Question {
	return &survey.Question{
		Name:     name,
		Prompt:   &survey.Password{Message: message},
		Validate: survey.Required,
	}
}

func (ui *UIService) getIntervalQn() *survey.Question {
	return &survey.Question{
		Name:   "interval",
		Prompt: &survey.Input{Message: "What is the monitoring interval (seconds)?"},
		Validate: func(val interface{}) error {
			strVal, ok := val.(string)
			if !ok {
				return errors.New("invalid interval")
			}

			intVal, err := strconv.Atoi(strVal)
			if err != nil {
				return errors.New("invalid interval")
			}

			if intVal <= 0 {
				return errors.New("the monitoring interval must be larger than zero seconds")
			}

			return nil
		},
	}
}
