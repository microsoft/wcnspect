package server

import (
	"bufio"
	"fmt"
	"io"
	"log"
	"os/exec"
	"strings"

	"github.com/microsoft/winspect/pkg/nets"
	pb "github.com/microsoft/winspect/rpc"
)

func addPktmonFilters(args map[string][]string) {
	// If empty, add an empty protocol since our filtering mechanism depends on there being one
	if len(args["protocols"]) == 0 {
		args["protocols"] = append(args["protocols"], "")
	}

	for _, protocol := range args["protocols"] {
		name := " winspect" + protocol + " "
		filters := []string{}

		// Build filter slice
		for arg, addrs := range args {
			// Short circuiting conditional for adding protocol(s) if in filter request
			if len(addrs) > 0 && len(addrs[0]) > 0 {
				filters = append(filters, pktParams[arg]+" "+strings.Join(addrs, " "))
			}
		}

		// If there are no filters, break
		if len(filters) == 0 {
			break
		}

		// Execute filter command
		fmt.Println("Applying filters...")
		if err := exec.Command("cmd", "/c", "pktmon filter add"+name+strings.Join(filters, " ")).Run(); err != nil {
			log.Printf("Failed to add%sfilter: %v", name, err)
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
