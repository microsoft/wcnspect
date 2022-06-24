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

var validProtocols = strings.Split(common.ValidProtocols, " ")

func ParseValidateNodes(nodes []string, nodeset []v1.Node) []string {
	winNodes := k8spi.FilterNodes(nodeset, k8spi.WindowsOS)
	winMap := k8spi.GetNodeMap(winNodes)
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

func ValidateCaptureParams(args *CaptureParams) {
	if err := ValidateTime(args.Time); err != nil {
		log.Fatal(err)
	}

	if err := ValidateIPAddrs(args.Ips); err != nil {
		log.Fatal(err)
	}

	if err := ValidateProtocols(args.Protocols); err != nil {
		log.Fatal(err)
	}

	if err := ValidatePorts(args.Ports); err != nil {
		log.Fatal(err)
	}

	if err := ValidateMACAddrs(args.Macs); err != nil {
		log.Fatal(err)
	}
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
		if !comprise.Contains(validProtocols, prot) {
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