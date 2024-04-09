package netutil

import (
	"net"
)

// GetOutboundIP gets a local IP address.
//
// Uses UDP instead of TCP because no handshake is required to establish a connection.
// Actually it does not establish any connection and the destination does not need to be existed at all.
//
// Source - https://stackoverflow.com/questions/23558425/how-do-i-get-the-local-ip-address-in-go/37382208#37382208
func GetOutboundIP() *net.IP {
	conn, err := net.Dial("udp", "1.1.1.1:80")
	if err != nil {
		return nil
	}
	defer conn.Close()

	localAddr := conn.LocalAddr().(*net.UDPAddr)

	return &localAddr.IP
}
