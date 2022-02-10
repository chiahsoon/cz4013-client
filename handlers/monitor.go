package handlers

import (
	"fmt"
	"net"

	"github.com/AlecAivazis/survey/v2"
	"github.com/chiahsoon/cz4013/client/api"
	"github.com/chiahsoon/cz4013/client/helpers"
	"github.com/chiahsoon/cz4013/client/models"
)

func HandleMonitor(action models.UserSelectedAction, conn *net.UDPConn) {
	if action != models.MonitorAction {
		return
	}

	userInput := api.MonitorReq{}
	subPrompt := helpers.GetSubPromptsForAction()[action]
	err := survey.Ask(subPrompt, &userInput)
	if err != nil {
		fmt.Println(err)
		return
	}

	req := api.Request{
		Method: api.MonitorMethodAPI,
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
		resp := &api.CallbackPayload{}
		err = codec.Decode(respData, &resp)
		if err != nil {
			return err
		}

		fmt.Printf("From %s:\n", addr)
		switch api.CallbackFunctionId(resp.FunctionId) {
		case api.UpdateCallback:
			fmt.Println(resp.Data)
		case api.StopMonitoringCallback:
			fmt.Println("Ending monitor interval ...")
			return nil
		default:
			fmt.Println("Invalid callback")
		}
	}
}
