package handlers

import (
	"net"

	"github.com/AlecAivazis/survey/v2"
	"github.com/chiahsoon/cz4013-client/api"
	apiModels "github.com/chiahsoon/cz4013-client/api/models"
	"github.com/chiahsoon/cz4013-client/models"
	"github.com/chiahsoon/cz4013-client/services"
)

func HandleMonitor(action models.UserSelectedAction, conn *net.UDPConn) {
	if action != models.MonitorAction {
		return
	}

	req := api.NewRequest()
	req.Method = string(api.MonitorAPI)
	input := apiModels.MonitorReq{}

	err := survey.Ask(services.UI.GetSubPromptsForAction()[action], &input)
	if err != nil {
		services.PP.PrintError(err.Error(), "", "")
		return
	}
	req.Data = input

	// Blocks while monitoring
	if err = listenForCallbacks(conn, &req); err != nil {
		services.PP.PrintError(err.Error(), "", "")
		return
	}
}

func listenForCallbacks(conn *net.UDPConn, req *api.Request) error {
	// Initiate monitoring
	codec := api.Codec{}
	encoded, err := codec.Encode(req)
	if err != nil {
		return err
	}

	// Initial request to start monitoring
	if err := services.ConnSvc.SendRequest(conn, encoded); err != nil {
		return err
	}

	// Listen for update callbacks
	for {
		resp := &apiModels.CallbackPayload{}
		if err := services.ConnSvc.GetResponse(conn, resp); err != nil {
			return err
		}

		switch apiModels.CallbackFunctionId(resp.FunctionId) {
		case apiModels.UpdateCallback:
			services.PP.Print(resp.Data, "- Updated Accounts -", "")
		case apiModels.StopMonitoringCallback:
			services.PP.PrintMessage("Ending monitor interval ...", "", "")
			return nil
		default:
			services.PP.PrintError("Invalid Callback", "", "")
		}
	}
}
