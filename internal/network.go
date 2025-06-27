package internal

import (
	"fmt"
	"net"
	"time"
)

// CheckHostPort checks if a host:port combination is reachable
// It will retry until the connection succeeds or timeout is reached
func CheckHostPort(host string, port int, msRetryTimeout int) bool {
	address := fmt.Sprintf("%s:%d", host, port)
	deadline := time.Now().Add(time.Duration(msRetryTimeout) * time.Millisecond)

	for time.Now().Before(deadline) {
		conn, err := net.DialTimeout("tcp", address, 500*time.Millisecond)
		if err == nil {
			_ = conn.Close()
			return true
		}
		time.Sleep(100 * time.Millisecond)
	}

	return false
}
