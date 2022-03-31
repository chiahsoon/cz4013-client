package handlers

import (
	"net"
	"time"

	"github.com/AlecAivazis/survey/v2"
	"github.com/chiahsoon/cz4013-client/api"
	"github.com/chiahsoon/cz4013-client/api/codec"
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

	// Initial request to start monitoring
	codec := &codec.Codec{}
	encoded, err := codec.Encode(req)
	if err != nil {
		services.PP.PrintError(err.Error(), "", "")
		return
	}

	if err := services.ConnSvc.SendRequest(conn, encoded); err != nil {
		services.PP.PrintError(err.Error(), "", "")
		return
	}

	// Block while monitoring
	intervalEnd := time.Now().Add(time.Duration(input.Interval) * time.Second)
	if err = listenForCallbacks(conn, intervalEnd); err != nil {
		services.PP.PrintError(err.Error(), "", "")
		return
	}

	services.PP.PrintMessage("Ending interval ...", "", "")
}

func listenForCallbacks(conn *net.UDPConn, intervalEnd time.Time) error {
	defer conn.SetDeadline(time.Time{}) // Reset to no deadlines after
	codec := codec.Codec{}

	for time.Now().Before(intervalEnd) {
		conn.SetDeadline(intervalEnd)
		resp := &api.Response{}
		if err := services.ConnSvc.GetResponse(conn, resp); err != nil {
			if err, ok := err.(net.Error); ok && err.Timeout() {
				return nil
			}
			return err
		}

		// Monitoring callbacks will always be string data
		var respData string
		if err := codec.DecodeAsInterface(resp.Data, &respData); err != nil {
			services.PP.PrintError(err.Error(), "", "")
			continue
		}

		services.PP.Print(respData, "", "")
	}

	return nil
}
