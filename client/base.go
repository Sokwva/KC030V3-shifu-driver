package client

import (
	"net"
	"time"
)

func Send(target string, cmd []byte) ([]byte, error) {
	tcpConn, err := net.DialTimeout("tcp", target, 5*time.Second)

	if err != nil {
		panic(err)
	}
	defer tcpConn.Close()

	tcpConn.Write(cmd)
	var buf []byte = make([]byte, 20)
	_, err = tcpConn.Read(buf)
	if err != nil {
		panic(err)
	}
	return buf, err
}
