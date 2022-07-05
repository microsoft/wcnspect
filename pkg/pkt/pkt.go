package pkt

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"os/exec"
	"strings"

	"github.com/microsoft/winspect/pkg/comprise"
	"github.com/microsoft/winspect/pkg/netutil"
	pb "github.com/microsoft/winspect/rpc"
)

var pktParams = map[string]string{
	"protocols": "-t",
	"ips":       "-i",
	"ports":     "-p",
	"macs":      "-m",
}

func AddFilters(filters *pb.Filters, verbose bool) error {
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

		if verbose {
			fmt.Println("Applying filters...")
		}

		filter := "pktmon filter add" + " " + name + " " + strings.Join(filterBuilder, " ")
		if err := exec.Command("cmd", "/c", filter).Run(); err != nil {
			return fmt.Errorf("failed to add %s filter: %v", name, err)
		}
	}

	return nil
}

func CreateStreamChannel(stdout *io.ReadCloser) <-chan string {
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

func ModifyCaptureCmd(mods *pb.Modifiers) (string, error) {
	baseCmd := "pktmon start -c -m real-time"
	pods, pktType, countersOnly := mods.GetPods(), mods.GetPacketType(), mods.GetCountersOnly()

	// Add packet type (all, flow, drop)
	baseCmd += fmt.Sprintf(" --type %s", pktType)

	// If we have pod IPs, then change the pktmonStartCommand
	if len(pods) > 0 {
		podIDs, err := netutil.GetPodIDs(pods)

		if err != nil {
			return "", err
		}

		baseCmd += fmt.Sprintf(" --comp %s", strings.Join(podIDs, " "))
	}

	// Specify whether cmd is counters only
	if countersOnly {
		baseCmd += " --counters-only"
	}

	return baseCmd, nil
}

func PullCounters() (string, error) {
	cmd := exec.Command("cmd", "/c", "pktmon stop")
	out, err := cmd.CombinedOutput()
	return string(out), err
}

func PullStreamCounters(includeHidden bool) (string, error) {
	pktmonCmd := "pktmon counter"

	if includeHidden {
		pktmonCmd += " --include-hidden"
	}

	cmd := exec.Command("cmd", "/c", pktmonCmd)
	out, err := cmd.Output()

	return string(out), err
}

func ResetFilters() error {
	if err := exec.Command("cmd", "/c", "pktmon filter remove").Run(); err != nil {
		return fmt.Errorf("failed to remove old filters: %v", err)
	}

	return nil
}

func ResetCaptureProgram() error {
	if err := exec.Command("cmd", "/c", "pktmon stop").Run(); err != nil {
		return fmt.Errorf("failed to stop pktmon: %v", err)
	}

	return nil
}

func StartStream(ctx context.Context, strCmd string) (*exec.Cmd, *io.ReadCloser, error) {
	if err := ResetCaptureProgram(); err != nil {
		return nil, nil, err
	}

	cmd := exec.CommandContext(ctx, "cmd", "/c", strCmd)
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return nil, nil, err
	}

	if err := cmd.Start(); err != nil {
		return nil, nil, err
	}

	return cmd, &stdout, nil
}
