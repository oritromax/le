package utils

import "net"


func GetLocalIP() (string, error) {
	// Connect to a dummy address; doesn't have to be reachable
	conn, err := net.Dial("udp", "8.8.8.8:80")
	if err != nil {
		return "", err
	}
	defer conn.Close()

	localAddr := conn.LocalAddr().(*net.UDPAddr)
	return localAddr.IP.String(), nil
}
