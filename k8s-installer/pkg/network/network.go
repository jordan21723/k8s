package network

import (
	"bytes"
	"errors"
	"fmt"
	"net"
	"os/exec"
	"strconv"
	"strings"
	"time"

	cmdLib "k8s-installer/pkg/command"
	"k8s-installer/pkg/log"
)

func OpenPort(ip net.IP, port int, zone string) (err error) {
	_, err = net.ListenTCP("tcp", &net.TCPAddr{
		IP:   ip,
		Port: port,
		Zone: zone,
	})
	return
}

func IntersectEachOther(cidrA, cidrB string) (bool, error) {
	ipA, ipNetA, errA := net.ParseCIDR(cidrA)
	if errA != nil {
		return true, errA
	}
	ipB, ipNetB, errB := net.ParseCIDR(cidrB)
	if errB != nil {
		return true, errB
	}
	return ipNetA.Contains(ipB) || ipNetB.Contains(ipA), nil
}

func OpenPortAndListen(ip net.IP, port int, zone string, stopChan <-chan struct{}) {
	ln, err := net.Listen("tcp", fmt.Sprintf(":%d", port))
	if err != nil {
		panic(err)
	}
	go func(ln net.Listener) {
		for {
			conn, _ := ln.Accept()
			conn.Close()
		}
	}(ln)

	<-stopChan
	ln.Close()
}

func GetDefaultGateway(ipv4 bool) (string, *net.Interface, error) {
	// run command "ip r | grep -i 'default via'"
	var command *exec.Cmd
	if ipv4 {
		command = exec.Command("ip", "r")
	} else {
		// might broken on grep version 2.20
		command = exec.Command("ip", "-6", "r")
	}
	cmdGrepDefaultGW := exec.Command("grep", "default via")
	output, stderr, err := cmdLib.CmdPipeline(command, cmdGrepDefaultGW)
	if err != nil {
		log.Debug(fmt.Sprintf("Failed to get default gateway due to error: \"%s\"", stderr))
		return "", nil, err
	}
	nicInfo := strings.Split(string(output), " ")
	if len(nicInfo) < 4 {
		return "", nil, errors.New(fmt.Sprintf("Failed to get gateway info by issue \"%s\"", command.String()))
	}
	interFaceName := nicInfo[4]
	defaultGw := nicInfo[2]
	nic, errNic := net.InterfaceByName(interFaceName)
	if errNic != nil {
		return "", nil, errNic
	}
	return defaultGw, nic, nil
}

func GetDefaultIP(ipv4 bool) net.IP {
	_, nic, err := GetDefaultGateway(ipv4)
	if err != nil {
		log.Fatal(err)
	}

	addresses, errIp := nic.Addrs()
	if errIp != nil {
		log.Fatal(errIp)
	}

	ip, _, errCidr := net.ParseCIDR(addresses[0].String())
	if errCidr != nil {
		log.Fatal(errCidr)
	}
	if ipv4 {
		return ip.To4()
	} else {
		return ip.To16()
	}
}

func GetNicWithCIDR(cidr string) (*net.Interface, error) {
	var stdOut, stdErr bytes.Buffer
	var err error
	interfaceName := ""
	ipv4 := true
	if ip, _, err := net.ParseCIDR(cidr); err != nil {
		return nil, err
	} else {
		ipv4 = !strings.Contains(ip.String(), ":")
	}
	if ipv4 {
		stdOut, stdErr, err = cmdLib.RunCmd("ip", "r")
	} else {
		stdOut, stdErr, err = cmdLib.RunCmd("ip", "-6", "r")
	}
	if err != nil {
		log.Error(fmt.Sprintf("Failed to get nic information by input CIDR \"%s\" due to error \"%s\" ", cidr, stdErr.String()))
		return nil, err
	}
	routes := strings.Split(stdOut.String(), "\n")
	for _, route := range routes {
		if strings.HasPrefix(route, cidr) {
			interfaceName = strings.Split(route, " ")[2]
			break
		}
	}
	if nic, err := net.InterfaceByName(interfaceName); err != nil {
		log.Error(fmt.Sprintf("Failed to get interface info with name \"%s\" ", interfaceName))
		return nil, err
	} else {
		return nic, err
	}
}

func CheckIpIsReachable(ip string, port int, ipFamily string, timeoutSec time.Duration) error {
	return CheckAddressIsReachable(ip+":"+strconv.Itoa(port), ipFamily, timeoutSec)
}

func CheckAddressIsReachable(address string, protocol string, timeoutSec time.Duration) error {
	connection, err := net.DialTimeout(protocol, address, timeoutSec)
	if err == nil {
		connection.Close()
	}
	return err
}

func GetAllIpAddress() ([]string, []string, error) {
	var v4Result, v6Result []string
	addresses, err := net.InterfaceAddrs()
	if err != nil {
		log.Errorf("Failed to get node all ip address due to error %s", err.Error())
	}
	for _, address := range addresses {
		isIpv4 := !strings.Contains(address.String(), ":")
		if isIpv4 {
			v4Result = append(v4Result, address.String())
		} else {
			v6Result = append(v6Result, address.String())
		}

	}
	return v4Result, v6Result, nil
}
