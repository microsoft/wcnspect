package nets

import (
	"bufio"
	"fmt"
	"log"
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
	return hcn.HostComputeEndpoint{}, fmt.Errorf("Endpoint with IP:%s not found", ip)
}

func GetPktmonID(mac string) (string, error) {
	out, err := exec.Command("cmd", "/c", "pktmon list").Output()
	if err != nil {
		log.Fatalf("Failed to run 'pktmon list': %v", err)
	}

	scanner := bufio.NewScanner(strings.NewReader(string(out)))
	for scanner.Scan() {
		line := scanner.Text()
		fields := strings.Fields(line)

		if len(fields) > 2 && fields[1] == mac {
			if scanner.Err() != nil {
				log.Print(err)
			}

			return fields[0], nil
		}
	}

	return "", fmt.Errorf("Packet Monitor component with MAC:%s not found", mac)
}

func GetPodIDs(pods []string) (ret []string) {
	endpoints, err := hcn.ListEndpoints()
	if err != nil {
		log.Fatal(err)
	}

	for _, pod := range pods {
		endpoint, err := GetEndpoint(endpoints, pod)
		if err != nil {
			log.Fatal(err)
		}

		id, err := GetPktmonID(endpoint.MacAddress)
		if err != nil {
			log.Fatal(err)
		}

		ret = append(ret, id)
	}

	return
}
