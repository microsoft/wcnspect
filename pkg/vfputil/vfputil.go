package vfputil

import (
	"fmt"
	"os/exec"
	"regexp"
	"strings"

	"github.com/microsoft/winspect/pkg/netutil"
)

type VFPPort struct {
	Name       string
	MacAddress string
	Type       string
}

func CollateCounters(pod string, verbose bool) (string, error) {
	var collated strings.Builder
	delim := "\n=========================================================================\n"

	pguid, err := GetPodPortGUID(pod)
	if err != nil {
		return collated.String(), err
	}

	guids := []string{pguid}
	titles := []string{fmt.Sprintf("Pod Port VFP Counters (ID: %s)", guids[0])}
	if verbose {
		hguid, eguid, err := GetHostAndExternalPortGUIDs()
		if err != nil {
			return collated.String(), err
		}

		guids = append(guids, hguid, eguid)
		titles = append(titles,
			fmt.Sprintf("Host vNic VFP Counters (ID: %s)", hguid),
			fmt.Sprintf("External Adapter VFP Counters (ID: %s)", eguid),
		)
	}

	collated.WriteString("\n")
	for i, guid := range guids {
		counters, err := PullVFPCounters(guid)
		if err != nil {
			return collated.String(), err
		}

		collated.WriteString(titles[i])
		collated.WriteString(delim)
		collated.WriteString(counters)

		if len(guids)-1 != i {
			collated.WriteString("\n\n")
		}
	}

	return collated.String(), nil
}

func GetPodPortGUID(podIP string) (string, error) {
	objs, err := netutil.ParseHNSDiag("endpoints")
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

/* Retrieves host port guid and external port guid on the VM running the function.
Returns host port guid and external port guid. Returns an error if there was an issue with retrieval.
*/
func GetHostAndExternalPortGUIDs() (string, string, error) {
	var hguid, eguid string

	objs, err := netutil.ParseHNSDiag("networks")
	if err != nil {
		return hguid, eguid, err
	}

	if len(objs) == 0 {
		return hguid, eguid, fmt.Errorf("no network objects found")
	}

	ip := objs[0].ManagementIP
	mac, err := PortIPtoMAC(ip)
	if err != nil {
		return hguid, eguid, err
	}

	vfpPorts, err := ParseVFPPorts()
	if err != nil {
		return hguid, eguid, err
	}

	for _, port := range vfpPorts {
		if port.MacAddress != mac {
			continue
		}

		switch port.Type {
		case "Internal":
			hguid = port.Name
		case "External":
			eguid = port.Name
		}
	}

	return hguid, eguid, nil
}

func ListVFPPorts() ([]byte, error) {
	return exec.Command("cmd", "/c", "vfpctrl /list-vmswitch-port").CombinedOutput()
}

func ParseVFPPorts() ([]VFPPort, error) {
	ret := []VFPPort{}

	out, err := ListVFPPorts()
	if err != nil {
		return ret, err
	}

	// Some very case-specific parsing
	delim := "\r\n"
	ports := string(out)
	ports = strings.ReplaceAll(ports, " ", "")
	for _, port := range strings.Split(ports, delim+delim) {
		if !strings.Contains(port, "Portname") || !strings.Contains(port, "MACaddress") {
			continue
		}

		p := VFPPort{}
		for _, line := range strings.Split(port, delim) {
			fields := strings.Split(line, ":")
			switch fields[0] {
			case "Portname":
				p.Name = fields[1]
			case "MACaddress":
				p.MacAddress = fields[1]
			case "Porttype":
				p.Type = fields[1]
			}
		}

		ret = append(ret, p)
	}

	return ret, nil
}

func PortIPtoMAC(ip string) (string, error) {
	out, err := netutil.ListIPConfig()
	if err != nil {
		return "", err
	}

	// Some very case-specific parsing
	delim := "\r\n"
	ipconfigs := string(out)
	ipconfigs = strings.ReplaceAll(ipconfigs, " ", "")
	for _, ipconfig := range strings.Split(ipconfigs, delim+delim) {
		if !strings.Contains(ipconfig, "IPv4Address") || !strings.Contains(ipconfig, "PhysicalAddres") {
			continue
		}

		var ipv4, mac string
		for _, line := range strings.Split(ipconfig, delim) {
			switch {
			case strings.Contains(line, "IPv4Address"):
				re := regexp.MustCompile(`(25[0-5]|2[0-4][0-9]|[01]?[0-9][0-9]?)(\.(25[0-5]|2[0-4][0-9]|[01]?[0-9][0-9]?)){3}`)
				ipv4 = re.FindString(line)
			case strings.Contains(line, "PhysicalAddres"):
				re := regexp.MustCompile(`((?:[\da-fA-F]{2}[:\-]){5}[\da-fA-F]{2})`)
				mac = re.FindString(line)
			}
		}

		if ip == ipv4 {
			return mac, nil
		}
	}

	return "", fmt.Errorf("unable to find corresponding MAC address for IP: %s", ip)
}

func PullVFPCounters(portGUID string) (string, error) {
	vfpCmd := fmt.Sprintf("vfpctrl /port %s /get-port-counter", portGUID)

	cmd := exec.Command("cmd", "/c", vfpCmd)
	out, err := cmd.Output()

	return string(out), err
}
