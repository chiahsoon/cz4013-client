package helpers

import (
	"errors"
	"strconv"

	"github.com/AlecAivazis/survey/v2"
	"github.com/chiahsoon/cz4013/client/models"
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
		models.OpenAccountAction:  {getNameQn(), getCurrencyQn()},
		models.CloseAccountAction: {getAccountNumberQn(), getNameQn()},
		models.GetBalanceAction:   {getAccountNumberQn(), getNameQn(), getCurrencyQn()},
		models.DepositAction:      {getAccountNumberQn(), getNameQn(), getCurrencyQn(), getAmountQn()},
		models.WithdrawAction:     {getAccountNumberQn(), getNameQn(), getCurrencyQn(), getAmountQn()},
		models.MonitorAction:      {getIntervalQn()},
	}
}

func GetPassword() *survey.Password {
	return &survey.Password{
		Message: "What is your password?",
	}
}

func getNameQn() *survey.Question {
	return makeQuestion("name", "What is your name?")
}

func getAccountNumberQn() *survey.Question {
	return makeQuestion("accountNumber", "What is your account number?")
}

func getCurrencyQn() *survey.Question {
	return makeQuestion("currency", "What is the currency?")
}

func getAmountQn() *survey.Question {
	return makeQuestion("amount", "What is the amount?")
}

func makeQuestion(name string, message string) *survey.Question {
	return &survey.Question{
		Name:     name,
		Prompt:   &survey.Input{Message: message},
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
