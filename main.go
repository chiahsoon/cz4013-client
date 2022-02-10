package main

import (
	"flag"
	"fmt"
	"net"

	"github.com/AlecAivazis/survey/v2"
	"github.com/chiahsoon/cz4013/client/handlers"
	"github.com/chiahsoon/cz4013/client/helpers"
	"github.com/chiahsoon/cz4013/client/models"
)

func connect(host string, port string) (*net.UDPConn, error) {
	addr := fmt.Sprintf("%s:%s", host, port)
	s, err := net.ResolveUDPAddr("udp4", addr)
	if err != nil {
		return nil, nil
	}

	conn, err := net.DialUDP("udp4", nil, s)
	if err != nil {
		return nil, nil
	}

	return conn, nil
}

func main() {
	host := flag.String("host", "localhost", "IP address of the server")
	port := flag.String("port", "5000", "Port of the server")
	flag.Parse()

	var err error
	conn, err := connect(*host, *port)
	if err != nil {
		fmt.Println(err)
		return
	}
	defer conn.Close()

	for {
		actionIdx := -1
		err = survey.AskOne(helpers.GetMainPrompt(), &actionIdx)
		if err != nil {
			fmt.Println(err)
			return
		}

		action := models.UserSelectedAction(actionIdx)
		if err := action.IsValid(); err != nil {
			fmt.Println("Exiting ...")
			return
		}

		handlers.HandleOpenAccount(action, conn)
		handlers.HandleCloseAccount(action, conn)
		handlers.HandleGetBalance(action, conn)
		handlers.HandleDeposit(action, conn)
		handlers.HandleWithdraw(action, conn)
		handlers.HandleMonitor(action, conn)
	}
}