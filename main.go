package main

import (
	"flag"
	"fmt"
	"net"
	"time"

	"github.com/AlecAivazis/survey/v2"
	"github.com/chiahsoon/cz4013-client/config"
	"github.com/chiahsoon/cz4013-client/handlers"
	"github.com/chiahsoon/cz4013-client/models"
	"github.com/chiahsoon/cz4013-client/services"
)

func connect(host string, port string) *net.UDPConn {
	addr := fmt.Sprintf("%s:%s", host, port)
	s, err := net.ResolveUDPAddr("udp4", addr)
	if err != nil {
		panic(err)
	}

	conn, err := net.DialUDP("udp4", nil, s)
	if err != nil {
		panic(err)
	}

	return conn
}

func main() {
	host := flag.String("host", "localhost", "IP address of the server")
	port := flag.String("port", "5000", "Port of the server")
	semantic := flag.String("semantic", string(config.AtLeastOnce), "Invocation Semantic - at-least-once (Default), at-most-once")
	flag.Parse()

	// Initialise command line configurations
	config.Global = &config.Config{}
	config.Global.InvocationSemantic = config.InvocationSemantic(*semantic)
	config.Global.Host = *host
	config.Global.Port = *port
	if err := config.Global.Validate(); err != nil {
		panic(err)
	}

	// Initialise server connection
	conn := connect(config.Global.Host, config.Global.Port)
	defer conn.Close()

	// Initialise services
	services.PP = &services.PrettyPrinter{}
	services.UI = &services.UIService{}
	services.ConnSvc = &services.ConnectionService{}
	services.ConnSvc.InvocationSemantic = config.Global.InvocationSemantic
	services.ConnSvc.TimeoutInterval = time.Duration(1) * time.Second
	services.ConnSvc.MaxRetryCount = -1

	// Handle user actions
	actionIdx := -1
	for {
		if err := survey.AskOne(services.UI.GetMainPrompt(), &actionIdx); err != nil {
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
		handlers.HandleCheckState(action, conn)
		handlers.HandleTransfer(action, conn)
	}
}
