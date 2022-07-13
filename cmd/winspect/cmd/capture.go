package cmd

import (
	"log"
	"os"
	"os/signal"
	"strings"
	"sync"
	"syscall"

	"github.com/microsoft/winspect/pkg/client"
	"github.com/microsoft/winspect/pkg/comprise"

	"github.com/spf13/cobra"
)

type captureCmd struct {
	time int32

	ips       []string
	protocols []string
	ports     []string
	macs      []string

	packetType   string
	countersOnly bool

	*baseBuilderCmd
}

func (b *commandsBuilder) newCaptureCmd() *captureCmd {
	cc := &captureCmd{}

	cmd := &cobra.Command{
		Use:   "capture",
		Short: "The 'capture' command will run a packet capture on all windows nodes.",
		Long: `The 'capture' command will run a packet capture on all windows nodes. For example:
	'winspect capture pods {pods} --protocols TCP -d 10'.`,
	}

	captureTypes := []string{"all", "nodes", "pods"}
	captureHelp := map[string]string{
		"all":   "Runs on all windows nodes in the AKS cluster.",
		"nodes": "Specify which nodes winspect should send requests to using node names.",
		"pods":  "Specify which pods the capture should filter on. Supports up to two pod names. Automatically defines nodes to capture on.",
	}
	for _, name := range captureTypes {
		subcmd := &cobra.Command{
			Use:   name,
			Short: captureHelp[name],
			Run: func(cmd *cobra.Command, args []string) {
				cc.printCapture(cmd.Name(), args)
			},
		}

		cmd.AddCommand(subcmd)
	}

	cmd.PersistentFlags().Int32VarP(&cc.time, "time", "d", 0, "Time to run packet capture for (in seconds). Runs indefinitely given 0.")

	cmd.PersistentFlags().StringSliceVarP(&cc.ips, "ips", "i", []string{}, "Match source or destination IP address. CIDR supported.")
	cmd.PersistentFlags().StringSliceVarP(&cc.protocols, "protocols", "t", []string{}, "Match by transport protocol. Can be TCP, UDP, ICMP, and/or TCP_{tcp flag}.")
	cmd.PersistentFlags().StringSliceVarP(&cc.ports, "ports", "r", []string{}, "Match source or destination port number.")
	cmd.PersistentFlags().StringSliceVarP(&cc.macs, "macs", "m", []string{}, "Match source or destination MAC address.")

	cmd.PersistentFlags().StringVar(&cc.packetType, "type", "all", "Select which packets to capture. Can be all, flow, or drop.")
	cmd.PersistentFlags().BoolVar(&cc.countersOnly, "counters-only", false, "Collect packet counters only. No packet logging.")

	cc.baseBuilderCmd = b.newBuilderCmd(cmd)

	return cc
}

func (cc *captureCmd) printCapture(subcmd string, endpoints []string) {
	targetNodes := cc.getTargetNodes()
	hostMap := make(map[string][]string)

	// Revise nodes and pods arguments based on command name
	switch subcmd {
	case "all":
		hostMap = comprise.CreateEmptyMap(targetNodes)
	case "nodes":
		if len(endpoints) == 0 {
			log.Fatal("must pass node names when using 'winspect capture nodes ...'")
		}

		nodes := strings.Split(endpoints[0], ",")
		targetNodes = client.ParseValidateNodes(nodes, cc.nodeSet)
		hostMap = comprise.CreateEmptyMap(targetNodes)
	case "pods":
		if len(endpoints) == 0 {
			log.Fatal("must pass pod names when using 'winspect capture pods ...'")
		}

		pods := strings.Split(endpoints[0], ",")
		hostMap = client.ParseValidatePods(pods, cc.podSet)
		targetNodes = comprise.Keys(hostMap)
	}

	// Capture any sigint to send a StopCapture request
	c := make(chan os.Signal)
	signal.Notify(c, os.Interrupt, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		<-c
		client.Cleanup(targetNodes)
		os.Exit(1)
	}()

	var wg sync.WaitGroup
	for _, ip := range targetNodes {
		wg.Add(1)

		c, closeClient := client.CreateConnection(ip)
		defer closeClient()

		captureArgs := &client.CaptureParams{
			Node:         ip,
			Pods:         hostMap[ip],
			Ips:          cc.ips,
			Protocols:    cc.protocols,
			Ports:        cc.ports,
			Macs:         cc.macs,
			Time:         cc.time,
			PacketType:   cc.packetType,
			CountersOnly: cc.countersOnly,
		}
		captureArgs.ValidateCaptureParams()

		go client.RunCaptureStream(c, captureArgs, &wg)
	}

	wg.Wait()
}
