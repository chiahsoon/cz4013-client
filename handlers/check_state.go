package handlers

import (
	"net"

	"github.com/chiahsoon/cz4013-client/api"
	"github.com/chiahsoon/cz4013-client/models"
	"github.com/chiahsoon/cz4013-client/services"
)

func HandleCheckState(action models.UserSelectedAction, conn *net.UDPConn) {
	if action != models.CheckStateAction {
		return
	}

	req := api.NewRequest()
	req.Method = string(api.CheckStateAPI)

	resp := api.Response{}
	err := services.ConnSvc.Fetch(conn, req, &resp)
	if err != nil {
		services.PP.PrintError(err.Error(), "", "")
		return
	}

	if resp.HasError() {
		services.PP.PrintError(resp.ErrMsg, "", "")
	} else {
		services.PP.Print(resp.Data, "- Response -", "")
	}
}
