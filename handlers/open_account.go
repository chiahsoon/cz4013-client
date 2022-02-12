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

func HandleOpenAccount(action models.UserSelectedAction, conn *net.UDPConn) {
	if action != models.OpenAccountAction {
		return
	}

	req := api.NewRequest()
	req.Method = string(api.OpenAccountAPI)
	input := apiModels.OpenAccountReq{}

	err := survey.Ask(helpers.GetSubPromptsForAction()[action], &input)
	if err != nil {
		fmt.Println(err)
		return
	}
	req.Data = input

	resp := api.Response{}
	err = helpers.Fetch(conn, req, &resp)
	if err != nil {
		fmt.Println(err)
		return
	}

	resp.Display()
}
