package helpers

import (
	"net"
	"time"

	"github.com/chiahsoon/cz4013-client/api"
	"github.com/chiahsoon/cz4013-client/config"
)

const (
	TimeoutInterval = time.Minute * time.Duration(1)
)

func Fetch(conn *net.UDPConn, req interface{}, dest interface{}) error {
	codec := api.Codec{}
	encoded, err := codec.Encode(req)
	if err != nil {
		return err
	}

	if config.Global.InvocationSemantic == config.Maybe {
		return fetch(conn, encoded, dest)
	}

	defer conn.SetDeadline(time.Time{}) // Reset to no timeout
	for {
		conn.SetDeadline(time.Now().Add(TimeoutInterval))
		if err := fetch(conn, encoded, dest); err != nil {
			continue
		}
		return nil
	}
}

func fetch(conn *net.UDPConn, reqData []byte, dest interface{}) error {
	if err := SendRequest(conn, reqData); err != nil {
		return err
	}

	return GetResponse(conn, dest)
}

func SendRequest(conn *net.UDPConn, reqData []byte) error {
	_, err := conn.Write(reqData)
	if err != nil {
		return err
	}
	return nil
}

func GetResponse(conn *net.UDPConn, dest interface{}) error {
	respData := make([]byte, 1024)
	n, _, err := conn.ReadFromUDP(respData)
	if err != nil {
		return err
	}

	respData = respData[0:n]
	codec := api.Codec{}
	return codec.Decode(respData, &dest)
}
