package main

import (
	"log"
	"net"
	"os"
	"strconv"
	"strings"
)

var validProtocols = []string{"TCP", "UDP", "ICMP", "ICMPv6"}
var vlog = log.New(os.Stderr, "", 0)

func contains(s []string, el string) bool {
	for _, value := range s {
		if value == el {
			return true
		}
	}
	return false
}

func validateTime(time int32) int32 {
	if time < 0 {
		vlog.Fatalf("Invalid time: %v", time)
	}
	return time
}

func parseValidateNodes(nodes string) []string {
	// Filters out any argument that isn't in the following format "<ip>:<port>"
	ls := strings.Split(nodes, ",")
	for _, node := range ls {
		if _, err := net.ResolveUDPAddr("udp", node); err != nil {
			vlog.Fatalf("Invalid node address: %v", err)
		}
	}
	return ls
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
		if !contains(validProtocols, prot) {
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
