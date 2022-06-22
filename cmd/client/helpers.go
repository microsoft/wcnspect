package main

import (
	"log"
	"net"
	"os"
	"strconv"
	"strings"

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

func parseValidateNodes(nodes string, nodeset []v1.Node) map[string][]string {
	winNodes := k8spi.FilterNodes(nodeset, k8spi.WindowsOS)
	winMap := k8spi.GetNodeMap(winNodes)
	winNames, winIPs := comprise.Keys(winMap), comprise.Values(winMap)

	if len(nodes) == 0 {
		return comprise.CreateEmptyMap(winIPs)
	}

	names := strings.Split(nodes, ",")
	for _, node := range names {
		if !comprise.Contains(winNames, node) {
			vlog.Fatalf("Invalid windows node name: %s", node)
		}
	}

	translateName := func(name string) string { return winMap[name] }
	return comprise.CreateEmptyMap(comprise.Map(names, translateName))
}

func parseValidatePods(pods string, podset []v1.Pod) map[string][]string {
	ret := make(map[string][]string)
	if len(pods) == 0 {
		return ret
	}

	podIPs, podNodes := k8spi.GetPodMaps(podset)
	podNames := comprise.Keys(podIPs)

	names := strings.Split(pods, ",")
	for _, pod := range names {
		if !comprise.Contains(podNames, pod) {
			vlog.Fatalf("Invalid pod name: %s", pod)
		}
	}

	for _, pod := range names {
		podIP, nodeIP := podIPs[pod], podNodes[pod]
		ret[nodeIP] = append(ret[nodeIP], podIP)
	}

	return ret
}

func parseValidateIPAddrs(ips string) []string {
	if len(ips) == 0 {
		return []string{}
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
		return []string{}
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
		return []string{}
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
		return []string{}
	}

	ls := strings.Split(macs, ",")
	for _, mac := range ls {
		if _, err := net.ParseMAC(mac); err != nil {
			vlog.Fatalf("Invalid MAC address: %v", err)
		}
	}
	return ls
}
