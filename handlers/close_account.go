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

func HandleCloseAccount(action models.UserSelectedAction, conn *net.UDPConn) {
	if action != models.CloseAccountAction {
		return
	}

	userInput := apiModels.CloseAccountReq{}
	subPrompt := helpers.GetSubPromptsForAction()[action]
	err := survey.Ask(subPrompt, &userInput)
	if err != nil {
		fmt.Println(err)
		return
	}

	err = survey.AskOne(helpers.GetPassword(), &userInput.Password)
	if err != nil {
		fmt.Println(err)
		return
	}

	req := api.Request{
		Method: string(api.CloseAccountAPI),
		Data:   userInput,
	}
	resp := api.Response{}
	err = helpers.Fetch(conn, req, &resp)
	if err != nil {
		return
	}

	resp.Display()
}
