package dali

import (
	"net"
	"os"
)

// Get local IPv4 address
func getLocalIPv4Address() (string, error) {
	host, err := os.Hostname()
	if err != nil {
		return "", err
	}

	addrs, err := net.LookupIP(host)
	if err != nil {
		return "", err
	}

	for _, addr := range addrs {
		if ipv4 := addr.To4(); ipv4 != nil {
			return ipv4.String(), nil
		}
	}

	return "", nil
}
