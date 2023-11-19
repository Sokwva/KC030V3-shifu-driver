package utils

import (
	"net"
	"time"
)

func ProbeTCP(target string) bool {
	conn, err := net.DialTimeout("tcp", target, 3*time.Second)
	if err != nil {
		return false
	}
	defer conn.Close()
	return true
}
