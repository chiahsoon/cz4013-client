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

func HandleMonitor(action models.UserSelectedAction, conn *net.UDPConn) {
	if action != models.MonitorAction {
		return
	}

	req := api.NewRequest()
	req.Method = string(api.MonitorAPI)
	input := apiModels.MonitorReq{}

	err := survey.Ask(helpers.GetSubPromptsForAction()[action], &input)
	if err != nil {
		fmt.Println(err)
		return
	}
	req.Data = input

	// Blocks while monitoring
	if err = listenForCallbacks(conn, &req); err != nil {
		fmt.Println(err)
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
	if err := helpers.SendRequest(conn, encoded); err != nil {
		return err
	}

	// Listen for update callbacks
	for {
		resp := &apiModels.CallbackPayload{}
		if err := helpers.GetResponse(conn, resp); err != nil {
			return err
		}

		fmt.Println("Update:")
		switch apiModels.CallbackFunctionId(resp.FunctionId) {
		case apiModels.UpdateCallback:
			fmt.Println(resp.Data)
		case apiModels.StopMonitoringCallback:
			fmt.Println("Ending monitor interval ...")
			return nil
		default:
			fmt.Println("Invalid callback")
		}
	}
}
