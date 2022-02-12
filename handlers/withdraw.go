package handlers

import (
	"fmt"
	"net"

	"github.com/AlecAivazis/survey/v2"
	"github.com/chiahsoon/cz4013-client/api"
	apiModels "github.com/chiahsoon/cz4013-client/api/models"
	"github.com/chiahsoon/cz4013-client/helpers"
	"github.com/chiahsoon/cz4013-client/models"
)

func HandleWithdraw(action models.UserSelectedAction, conn *net.UDPConn) {
	if action != models.WithdrawAction {
		return
	}

	req := api.NewRequest()
	req.Method = string(api.OpenAccountAPI)
	input := apiModels.UpdateBalanceReq{}

	err := survey.Ask(helpers.GetSubPromptsForAction()[action], &input)
	if err != nil {
		fmt.Println(err)
		return
	}
	input.Amount *= -1
	req.Data = input

	resp := api.Response{}
	err = helpers.Fetch(conn, req, &resp)
	if err != nil {
		fmt.Println(err)
		return
	}

	resp.Display()
}
