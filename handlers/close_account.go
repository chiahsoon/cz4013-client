package handlers

import (
	"net"

	"github.com/AlecAivazis/survey/v2"
	"github.com/chiahsoon/cz4013-client/api"
	"github.com/chiahsoon/cz4013-client/api/codec"
	apiModels "github.com/chiahsoon/cz4013-client/api/models"
	"github.com/chiahsoon/cz4013-client/models"
	"github.com/chiahsoon/cz4013-client/services"
)

func HandleCloseAccount(action models.UserSelectedAction, conn *net.UDPConn) {
	if action != models.CloseAccountAction {
		return
	}

	req := api.NewRequest()
	req.Method = string(api.CloseAccountAPI)
	input := apiModels.CloseAccountReq{}

	err := survey.Ask(services.UI.GetSubPromptsForAction()[action], &input)
	if err != nil {
		services.PP.PrintError(err.Error(), "", "")
		return
	}
	req.Data = input

	resp := api.Response{}
	err = services.ConnSvc.Fetch(conn, req, &resp)
	if err != nil {
		services.PP.PrintError(err.Error(), "", "")
		return
	}

	if resp.HasError() {
		services.PP.PrintError(resp.ErrMsg, "", "")
	} else {
		// Unmarshall and print rest of data
		c := codec.Codec{}
		var respData apiModels.CloseAccountResp
		if err := c.DecodeAsInterface(resp.Data, &respData); err != nil {
			services.PP.PrintError(err.Error(), "", "")
			return
		}

		services.PP.Print(respData, "- Response -", "")
	}
}
