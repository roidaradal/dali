package dali

import (
	"net"
	"os"
)

var (
	DiscoveryPort uint16 = 45678 // UDP discovery port
	TransferPort  uint16 = 45679 // TCP transfer port
)

type Peer struct {
	Name string
	Addr string
}

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
