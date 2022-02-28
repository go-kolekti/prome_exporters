package internal

import (
	"net"

	"trellis.tech/trellis/common.v1/shell"
)

// GetIP todo ipv6
func GetIP() string {
	ip := shell.Output("hostname -i")
	if ip != "" {
		return ip
	}

	addrs, err := net.InterfaceAddrs()
	if err != nil {
		return ""
	}
	for _, address := range addrs {
		if ipnet, ok := address.(*net.IPNet); ok && !ipnet.IP.IsLoopback() {
			if ipnet.IP.To4() != nil {
				return ipnet.IP.String()
			}
		}
	}
	return ""
}
