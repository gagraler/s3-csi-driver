package utils

import (
	"fmt"
	netutil "k8s.io/utils/net"
)

/**
 * @author: HuaiAn xu
 * @date: 2024-03-12 17:03:47
 * @file: net.go
 * @description: net utils
 */

// GetServerFromSource returns the server address from the source.
func GetServerFromSource(server string) string {
	if netutil.IsIPv6String(server) {
		return fmt.Sprintf("[%s]", server)
	}
	return server
}