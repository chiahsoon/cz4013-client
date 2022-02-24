package services

import (
	"errors"
	"net"
	"time"

	"github.com/chiahsoon/cz4013-client/api/codec"
	"github.com/chiahsoon/cz4013-client/config"
)

type ConnectionService struct {
	config.InvocationSemantic
	TimeoutInterval time.Duration
	MaxRetryCount   int
}

func (cs *ConnectionService) Fetch(conn *net.UDPConn, req interface{}, dest interface{}) error {
	// Assumes dest is already a pointer
	c := codec.Codec{}
	encoded, err := c.Encode(req)
	if err != nil {
		return err
	}
	if cs.InvocationSemantic == config.Maybe {
		return cs.fetch(conn, encoded, dest)
	}

	defer conn.SetDeadline(time.Time{}) // Reset to no timeout
	for i := 0; i < cs.MaxRetryCount; i++ {
		conn.SetDeadline(time.Now().Add(cs.TimeoutInterval))
		if err := cs.fetch(conn, encoded, dest); err != nil {
			continue
		}
		return nil
	}

	return errors.New("failed to get response")
}

func (cs *ConnectionService) SendRequest(conn *net.UDPConn, reqData []byte) error {
	_, err := conn.Write(reqData)
	if err != nil {
		return err
	}
	return nil
}

func (cs *ConnectionService) GetResponse(conn *net.UDPConn, dest interface{}) error {
	// Assumes dest is already a pointer
	respData := make([]byte, 1024)
	n, _, err := conn.ReadFromUDP(respData)
	if err != nil {
		return err
	}
	respData = respData[0:n]
	c := codec.Codec{}
	err = c.Decode(respData, dest)
	return err
}

func (cs *ConnectionService) fetch(conn *net.UDPConn, reqData []byte, dest interface{}) error {
	// Assumes dest is already a pointer
	if err := cs.SendRequest(conn, reqData); err != nil {
		return err
	}

	return cs.GetResponse(conn, dest)
}
