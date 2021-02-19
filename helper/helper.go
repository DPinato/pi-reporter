package helper

import (
	"net"
	"strings"
)

var PINetIfaces = []string{"eth0", "wlan0"}

// GetPIName returns the name of the current PI using the interface name specified
// name will be something like pi-<mac>
func GetPIName(ifName string) (string, error) {
	outputStr := "pi-"
	iface, err := net.InterfaceByName(ifName)
	if err != nil {
		return "", err
	}

	macStr := iface.HardwareAddr.String()
	macStr = strings.ReplaceAll(macStr, ":", "")
	outputStr += macStr
	return outputStr, err
}
