package helpers

import (
	"errors"
	"strconv"

	"github.com/AlecAivazis/survey/v2"
	"github.com/chiahsoon/cz4013-client/models"
)

func GetMainPrompt() *survey.Select {
	options := []string{}
	for i := 0; i < len(models.AllActions); i++ {
		options = append(options, models.AllActions[i].Description())
	}

	return &survey.Select{
		Message: "What would you like to do?",
		Options: append(options, "Exit"),
	}
}

func GetSubPromptsForAction() map[models.UserSelectedAction][]*survey.Question {
	return map[models.UserSelectedAction][]*survey.Question{
		models.OpenAccountAction:  {getNameQn(), getCurrencyQn(), getPasswordQn()},
		models.CloseAccountAction: {getAccountNumberQn(), getNameQn(), getPasswordQn()},
		models.GetBalanceAction:   {getAccountNumberQn(), getNameQn(), getCurrencyQn(), getPasswordQn()},
		models.DepositAction:      {getAccountNumberQn(), getNameQn(), getCurrencyQn(), getAmountQn(), getPasswordQn()},
		models.WithdrawAction:     {getAccountNumberQn(), getNameQn(), getCurrencyQn(), getAmountQn(), getPasswordQn()},
		models.MonitorAction:      {getIntervalQn()},
	}
}

func getPasswordQn() *survey.Question {
	return makePasswordQuestion("password", "What is your password?")
}

func getNameQn() *survey.Question {
	return makeTextQuestion("name", "What is your name?")
}

func getAccountNumberQn() *survey.Question {
	return makeTextQuestion("accountNumber", "What is your account number?")
}

func getCurrencyQn() *survey.Question {
	return makeTextQuestion("currency", "What is the currency?")
}

func getAmountQn() *survey.Question {
	return makeTextQuestion("amount", "What is the amount?")
}

func makeTextQuestion(name string, message string) *survey.Question {
	return &survey.Question{
		Name:     name,
		Prompt:   &survey.Input{Message: message},
		Validate: survey.Required,
	}
}

func makePasswordQuestion(name string, message string) *survey.Question {
	return &survey.Question{
		Name:     name,
		Prompt:   &survey.Password{Message: message},
		Validate: survey.Required,
	}
}

func getIntervalQn() *survey.Question {
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
