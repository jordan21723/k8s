package netutils

import (
	"fmt"
	"net"
	"strings"
	"time"

	"k8s-installer/pkg/constants"
)

const CoreDNSPort = 53

func CalculateTimeout(fileLength int64, minRate, defaultMinRate constants.Size, reservedTime time.Duration) time.Duration {
	// ensure the minRate to avoid trigger panic when minRate equals zero
	if fileLength <= 0 ||
		(minRate <= 0 && defaultMinRate <= 0) {
		return 0
	}
	if minRate <= 0 {
		minRate = defaultMinRate
	}

	return time.Duration(fileLength/int64(minRate))*time.Second + reservedTime
}

/**
 * Cut and assemble CoreDNS IP Addr from Kubernetes service CIDR.
 */
func GetCoreDNSAddr(serviceV4CIDR string) (string, error) {
	ip, _, err := net.ParseCIDR(serviceV4CIDR)
	if err != nil {
		return "", err
	}
	return ip.String()[:strings.LastIndex(ip.String(), ".")+1] + fmt.Sprintf("10:%d", CoreDNSPort), nil
}
