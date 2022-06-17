package main

import (
	"log"
	"net"
	"os"
	"strconv"
	"strings"

	"github.com/microsoft/winspect/common"
	"github.com/microsoft/winspect/pkg/comprise"
	"github.com/microsoft/winspect/pkg/k8spi"

	v1 "k8s.io/api/core/v1"
)

var validProtocols = []string{"TCP", "UDP", "ICMP", "ICMPv6"}
var vlog = log.New(os.Stderr, "", 0)

func validateTime(time int32) int32 {
	if time < 0 {
		vlog.Fatalf("Invalid time: %v", time)
	}
	return time
}

func parseValidateNodes(nodes string, nodeset []v1.Node) []string {
	winNodes := k8spi.FilterNodes(nodeset, k8spi.WindowsOS)
	winMap := k8spi.GetNodeMap(winNodes)
	winNames, winIPs := comprise.Keys(winMap), comprise.Values(winMap)
	addPort := func(s string) string { return s + ":" + common.DefaultServerPort }

	if len(nodes) == 0 {
		return comprise.Map(winIPs, addPort)
	}

	names := strings.Split(nodes, ",")
	for _, node := range names {
		if !comprise.Contains(winNames, node) {
			vlog.Fatalf("Invalid windows node name: %s", node)
		}
	}

	translateName := func(name string) string { return addPort(winMap[name]) }
	return comprise.Map(names, translateName)
}

func parseValidateIPAddrs(ips string) []string {
	if len(ips) == 0 {
		return []string{""}
	}

	ls := strings.Split(ips, ",")
	for _, ip := range ls {
		if _, _, err := net.ParseCIDR(ip); net.ParseIP(ip) == nil && err != nil {
			vlog.Fatalf("Invalid IP address: %s", ip)
		}
	}
	return ls
}

func parseValidateProts(protocols string) []string {
	if len(protocols) == 0 {
		return []string{""}
	}

	ls := strings.Split(protocols, ",")
	for _, prot := range ls {
		if !comprise.Contains(validProtocols, prot) {
			vlog.Fatalf("Invalid protocol: %s", prot)
		}
	}
	return ls
}

func parseValidatePorts(ports string) []string {
	if len(ports) == 0 {
		return []string{""}
	}

	ls := strings.Split(ports, ",")
	for _, port := range ls {
		if p, err := strconv.Atoi(port); (err != nil) || !(0 <= p && p <= 65535) {
			vlog.Fatalf("Invalid port: %s", port)
		}

	}
	return ls
}

func parseValidateMACAddrs(macs string) []string {
	if len(macs) == 0 {
		return []string{""}
	}

	ls := strings.Split(macs, ",")
	for _, mac := range ls {
		if _, err := net.ParseMAC(mac); err != nil {
			vlog.Fatalf("Invalid MAC address: %v", err)
		}
	}
	return ls
}
