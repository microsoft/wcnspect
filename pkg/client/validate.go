package client

import (
	"fmt"
	"log"
	"net"
	"strconv"
	"strings"

	"github.com/microsoft/winspect/common"
	"github.com/microsoft/winspect/pkg/comprise"
	"github.com/microsoft/winspect/pkg/k8spi"

	v1 "k8s.io/api/core/v1"
)

var validTCPFormats = comprise.Map(strings.Split(common.ValidTCPFlags, " "), func(s string) string { return "TCP_" + s })
var validProtocols = append(strings.Split(common.ValidProtocols, " "), validTCPFormats...)
var validPktTypes = strings.Split(common.ValidPacketTypes, " ")

func ValidateNodes(nodes []string, winNodes []string) error {
	for _, node := range nodes {
		if !comprise.Contains(winNodes, node) {
			return fmt.Errorf("invalid windows node name: %s", node)
		}
	}
	return nil
}

func ValidatePods(pods []string, podNames []string) error {
	for _, pod := range pods {
		if !comprise.Contains(podNames, pod) {
			return fmt.Errorf("invalid windows pod name: %s", pod)
		}
	}
	return nil
}

func ParseValidateNodes(nodes []string, nodeset []v1.Node) []string {
	winNodes := k8spi.FilterNodes(nodeset, k8spi.WindowsOS)
	winMap := k8spi.GetNodesNameToIp(winNodes)
	winNames, winIPs := comprise.Keys(winMap), comprise.Values(winMap)

	if len(nodes) == 0 {
		return winIPs
	}

	for _, node := range nodes {
		if !comprise.Contains(winNames, node) {
			log.Fatalf("invalid windows node name: %s", node)
		}
	}

	translateName := func(name string) string { return winMap[name] }
	return comprise.Map(nodes, translateName)
}

func ParseValidatePods(pods []string, podset []v1.Pod) map[string][]string {
	ret := make(map[string][]string)
	if len(pods) == 0 {
		return ret
	}

	podIPs, podNodes := k8spi.GetPodMaps(podset)
	podNames := comprise.Keys(podIPs)

	for _, pod := range pods {
		if !comprise.Contains(podNames, pod) {
			log.Fatalf("invalid pod name: %s", pod)
		}
	}

	for _, pod := range pods {
		podIP, nodeIP := podIPs[pod], podNodes[pod]
		ret[nodeIP] = append(ret[nodeIP], podIP)
	}

	return ret
}

func ValidateTime(time int32) error {
	if time < 0 {
		return fmt.Errorf("time should be greater than 0")
	}
	return nil
}

func ValidateIPAddrs(ips []string) error {
	for _, ip := range ips {
		if _, _, err := net.ParseCIDR(ip); net.ParseIP(ip) == nil && err != nil {
			return fmt.Errorf("invalid IP address format: %s", ip)
		}
	}
	return nil
}

func ValidateProtocols(protocols []string) error {
	for _, prot := range protocols {
		if !comprise.Contains(validProtocols, strings.ToUpper(prot)) {
			return fmt.Errorf("invalid protocol: %s", prot)
		}
	}
	return nil
}

func ValidatePorts(ports []string) error {
	for _, port := range ports {
		if p, err := strconv.Atoi(port); (err != nil) || !(0 <= p && p <= 65535) {
			return fmt.Errorf("invalid port number: %s", port)
		}

	}
	return nil
}

func ValidateMACAddrs(macs []string) error {
	for _, mac := range macs {
		if _, err := net.ParseMAC(mac); err != nil {
			return fmt.Errorf("invalid MAC address format: %v", err)
		}
	}
	return nil
}

func ValidatePktType(pktType string) error {
	if !comprise.Contains(validPktTypes, strings.ToUpper(pktType)) {
		return fmt.Errorf("invalid packet type: %s", pktType)
	}

	return nil
}
