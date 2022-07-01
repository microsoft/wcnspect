package server

import (
	"bufio"
	"fmt"
	"io"
	"log"
	"os/exec"
	"strings"

	"github.com/microsoft/winspect/pkg/comprise"
	"github.com/microsoft/winspect/pkg/nets"
	pb "github.com/microsoft/winspect/rpc"
)

var pktParams = map[string]string{
	"protocols": "-t",
	"ips":       "-i",
	"ports":     "-p",
	"macs":      "-m",
}

func addPktmonFilters(filters *pb.Filters) {
	protocols := filters.GetProtocols()
	args := map[string][]string{
		"ips":   filters.GetIps(),
		"ports": filters.GetPorts(),
		"macs":  filters.GetMacs(),
	}

	// If empty, add an empty protocol since our filtering mechanism depends on there being one
	if len(protocols) == 0 {
		protocols = append(protocols, "")
	}

	// Parse any tcp flags for valid pktmon filter input
	formatTCPFlags := func(s string) string { return strings.Replace(s, "_", " ", 1) }
	pktProtocols := comprise.Map(protocols, formatTCPFlags)

	for i, protocol := range protocols {
		name := "winspect" + strings.ToUpper(protocol)
		filterBuilder := []string{}

		// Build filter slice
		for arg, addrs := range args {
			if len(addrs) > 0 {
				newFilterFlag := pktParams[arg] + " " + strings.Join(addrs, " ")
				filterBuilder = append(filterBuilder, newFilterFlag)
			}
		}

		if len(protocol) > 0 {
			filterBuilder = append(filterBuilder, pktParams["protocols"]+" "+pktProtocols[i])
		}

		// If no filters, continue
		if len(filterBuilder) == 0 {
			continue
		}

		fmt.Println("Applying filters...")
		filter := "pktmon filter add" + " " + name + " " + strings.Join(filterBuilder, " ")
		if err := exec.Command("cmd", "/c", filter).Run(); err != nil {
			log.Printf("Failed to add %s filter: %v", name, err)
		}
	}
}

func pktmonStream(stdout *io.ReadCloser) <-chan string {
	c := make(chan string)

	scanner := bufio.NewScanner(*stdout)
	scanner.Split(bufio.ScanLines)
	go func(s *bufio.Scanner) {
		for s.Scan() {
			c <- s.Text()
		}
	}(scanner)

	return c
}

func pullCounters() string {
	cmd := exec.Command("cmd", "/c", "pktmon stop")

	out, err := cmd.Output()
	if err != nil {
		log.Print(err)
	}

	return string(out)
}

func resetCaptureContext(s *CaptureServer) {
	s.currMonitor = nil
	s.pktContextCancel = nil
	s.printCounters = false
}

func resetPktmon(captures bool, filters bool) error {
	// Stop pktmon
	if captures {
		if err := exec.Command("cmd", "/c", "pktmon stop").Run(); err != nil {
			log.Printf("Failed to stop pktmon: %v", err)
		}
	}

	// Clear filters
	if filters {
		if err := exec.Command("cmd", "/c", "pktmon filter remove").Run(); err != nil {
			log.Printf("Failed to remove old filters: %v", err)
		}
	}

	return nil
}

func revisePktmonCommand(mods *pb.Modifiers, cmd string) string {
	pods, pktType, countersOnly := mods.GetPods(), mods.GetPacketType(), mods.GetCountersOnly()

	// Add packet type (all, flow, drop)
	cmd += fmt.Sprintf(" --type %s", pktType)

	// If we have pod IPs, then change the pktmonStartCommand
	if len(pods) > 0 {
		podIDs := nets.GetPodIDs(pods)
		cmd += fmt.Sprintf(" --comp %s", strings.Join(podIDs, " "))
	}

	// Specify whether cmd is counters only
	if countersOnly {
		cmd += " --counters-only"
	}

	return cmd
}
