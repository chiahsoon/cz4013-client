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

	userInput := apiModels.MonitorReq{}
	subPrompt := helpers.GetSubPromptsForAction()[action]
	err := survey.Ask(subPrompt, &userInput)
	if err != nil {
		fmt.Println(err)
		return
	}

	req := api.Request{
		Method: string(api.MonitorAPI),
		Data:   userInput,
	}

	// Simulate "blocking" while monitoring
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

	_, err = conn.Write(encoded)
	if err != nil {
		return err
	}

	// Listen for update callbacks
	respData := make([]byte, 1024)
	for {
		n, addr, err := conn.ReadFromUDP(respData)
		if err != nil {
			return err
		}

		respData = respData[0:n]
		resp := &apiModels.CallbackPayload{}
		err = codec.Decode(respData, &resp)
		if err != nil {
			return err
		}

		fmt.Printf("From %s:\n", addr)
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
