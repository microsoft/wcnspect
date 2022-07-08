package vfputil

import (
	"fmt"
	"os/exec"

	"github.com/microsoft/winspect/pkg/netutil"
)

func GetPodPortGUID(podIP string) (string, error) {
	objs, err := netutil.ParseLogs()
	if err != nil {
		return "", err
	}

	for _, obj := range objs {
		if obj.IPAddress == podIP {
			allocators := obj.Resources.Allocators
			if len(allocators) == 0 {
				return "", fmt.Errorf("could not find Allocators for endpoint with IP: %s", podIP)
			}

			guid := allocators[0].EndpointPortGuid
			if guid == "" {
				return "", fmt.Errorf("PortGUID is empty for endpoint with IP: %s", podIP)
			}

			return allocators[0].EndpointPortGuid, nil
		}
	}

	return "", fmt.Errorf("endpoint with IP: %s not found", podIP)
}

func GetExternalPortGUID(podIP string) (string, error) {
	return "", nil //TODO:
}

func GetHostPortGUID(podIP string) (string, error) {
	return "", nil //TODO:
}

func PullVFPCounters(portGUID string) (string, error) {
	vfpCmd := fmt.Sprintf("vfpctrl /port %s /get-port-counter", portGUID)

	cmd := exec.Command("cmd", "/c", vfpCmd)
	out, err := cmd.Output()

	return string(out), err
}
