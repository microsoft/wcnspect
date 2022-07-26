// Copyright (c) Microsoft Corporation.
// Licensed under the MIT license.

package client

import (
	"fmt"
	"net"
	"strconv"
	"strings"

	"github.com/microsoft/winspect/common"
	"github.com/microsoft/winspect/pkg/comprise"
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
