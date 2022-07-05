package netutil

import (
	"bufio"
	"fmt"
	"os/exec"
	"strings"

	"github.com/Microsoft/hcsshim/hcn"
)

func GetEndpoint(endpoints []hcn.HostComputeEndpoint, ip string) (hcn.HostComputeEndpoint, error) {
	for _, endpoint := range endpoints {
		for _, ipconfig := range endpoint.IpConfigurations {
			if ipconfig.IpAddress == ip {
				return endpoint, nil
			}
		}
	}
	return hcn.HostComputeEndpoint{}, fmt.Errorf("endpoint with IP:%s not found", ip)
}

func GetLogs(option string, json bool) ([]byte, error) {
	cmd := fmt.Sprintf("hnsdiag list %s", option)

	if json {
		cmd += " -d"
	}

	return exec.Command("cmd", "/c", cmd).CombinedOutput()
}

func GetPktmonID(mac string) (string, error) {
	out, err := exec.Command("cmd", "/c", "pktmon list").Output()
	if err != nil {
		return "", fmt.Errorf("failed to run 'pktmon list': %v", err)
	}

	scanner := bufio.NewScanner(strings.NewReader(string(out)))
	for scanner.Scan() {
		line := scanner.Text()
		fields := strings.Fields(line)

		if len(fields) > 2 && fields[1] == mac {
			if scanner.Err() != nil {
				return "", err
			}

			return fields[0], nil
		}
	}

	return "", fmt.Errorf("packet monitor component with MAC:%s not found", mac)
}

/* Retrieves pktmon component vNic ID for each pod IP passed.
Returns string slice of these ids.
*/
func GetPodIDs(pods []string) (ret []string, err error) {
	var endpoints []hcn.HostComputeEndpoint
	var endpoint hcn.HostComputeEndpoint
	var id string

	endpoints, err = hcn.ListEndpoints()
	if err != nil {
		return
	}

	for _, pod := range pods {
		endpoint, err = GetEndpoint(endpoints, pod)
		if err != nil {
			return
		}

		id, err = GetPktmonID(endpoint.MacAddress)
		if err != nil {
			return
		}

		ret = append(ret, id)
	}

	return
}
