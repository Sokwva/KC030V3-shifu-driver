package client

import (
	"net"
	"slices"
	"time"
)

func Send(target string, cmd []byte) ([]byte, error) {
	tcpConn, err := net.DialTimeout("tcp", target, 5*time.Second)

	if err != nil {
		return []byte{}, err
	}
	defer tcpConn.Close()

	slices.Reverse(cmd)
	tcpConn.Write(cmd)
	var buf []byte = make([]byte, 20)
	_, err = tcpConn.Read(buf)
	if err != nil {
		return []byte{}, err
	}
	return buf, err
}
