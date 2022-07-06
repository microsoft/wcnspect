package netutil

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os/exec"
	"regexp"
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
	return hcn.HostComputeEndpoint{}, fmt.Errorf("endpoint with IP: %s not found", ip)
}

func GetLogs(option string, verbose bool) ([]byte, error) {
	cmd := fmt.Sprintf("hnsdiag list %s", option)

	if verbose {
		cmd += " -d"
	}

	return exec.Command("cmd", "/c", cmd).CombinedOutput()
}

func GetLogsMaps() ([]map[string]interface{}, error) {
	var ret []map[string]interface{}

	// Get logs
	logs, err := GetLogs("all", true)
	if err != nil {
		return ret, err
	}

	// Must modify logs string in order to parse as json
	re := regexp.MustCompile(`\n{`)
	temp := "[" + re.ReplaceAllString(string(logs), "\n,{") + "]"

	// Unmarshal into map
	err = json.Unmarshal([]byte(temp), &ret)

	return ret, err
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

func GetPortGUID(podIP string) (string, error) {
	endpoints, err := GetLogsMaps()
	if err != nil {
		return "", err
	}

	for _, m := range endpoints {
		if m["IpAddress"] == podIP {
			resources := m["Resources"]
			if resources == nil {
				return "", fmt.Errorf("could not find Resources for endpoint with IP: %s", podIP)
			}

			allocators := m["Allocators"]
			if allocators == nil {
				return "", fmt.Errorf("could not find Allocators for endpoint with IP: %s", podIP)
			}

			guid, ok := m["EndpointPortGUID"].(string)

			if !ok {
				return "", fmt.Errorf("not a string -> %#v", guid)
			}

			if guid == "" {
				return "", fmt.Errorf("PortGUID is empty for endpoint with IP: %s", podIP)
			}

			return guid, nil
		}
	}

	return "", fmt.Errorf("endpoint with IP: %s not found", podIP)
}

func GetPortGUIDs(pods []string) (ret []string, err error) {
	var guid string

	for _, pod := range pods {
		guid, err = GetPortGUID(pod)
		if err != nil {
			return
		}

		ret = append(ret, guid)
	}

	return
}
