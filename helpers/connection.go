package helpers

import (
	"net"

	"github.com/chiahsoon/cz4013/client/api"
)

func Fetch(conn *net.UDPConn, req interface{}, dest interface{}) error {
	codec := api.Codec{}
	encoded, err := codec.Encode(req)
	if err != nil {
		return err
	}

	_, err = conn.Write(encoded)
	if err != nil {
		return err
	}

	respData := make([]byte, 1024)
	n, _, err := conn.ReadFromUDP(respData)
	if err != nil {
		return err
	}

	respData = respData[0:n]
	err = codec.Decode(respData, &dest)
	if err != nil {
		return err
	}

	return nil
}
