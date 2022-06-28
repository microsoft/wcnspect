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

var captureArgs = client.CaptureParams{}

var captureCmd = &cobra.Command{
	Use:   "capture",
	Short: "The 'capture' command will run a packet capture on all windows nodes.",
	Long: `The 'capture' command will run a packet capture on all windows nodes. For example:
'winspect capture pods {pods} --protocols TCP -d 10'.`,
}

func init() {
	rootCmd.AddCommand(captureCmd)

	captureCmd.PersistentFlags().StringSliceVarP(&captureArgs.Ips, "ips", "i", []string{}, "Match source or destination IP address. CIDR supported.")
	captureCmd.PersistentFlags().StringSliceVarP(&captureArgs.Protocols, "protocols", "t", []string{}, "Match by transport protocol. Can be TCP, UDP, and/or ICMP).")
	captureCmd.PersistentFlags().StringSliceVarP(&captureArgs.Ports, "ports", "r", []string{}, "Match source or destination port number.")
	captureCmd.PersistentFlags().StringSliceVarP(&captureArgs.Macs, "macs", "m", []string{}, "Match source or destination MAC address.")
	captureCmd.PersistentFlags().Int32VarP(&captureArgs.Time, "time", "d", 0, "Time to run packet capture for (in seconds). Runs indefinitely given 0.")
	captureCmd.PersistentFlags().StringVar(&captureArgs.PacketType, "type", "all", "Select which packets to capture. Can be all, flow, or drop.")
	captureCmd.PersistentFlags().BoolVar(&captureArgs.CountersOnly, "counters-only", false, "Collect packet counters only. No packet logging.")

	captureTypes := []string{"all", "nodes", "pods"}
	captureHelp := map[string]string{
		"all":   "Runs on all windows nodes in the AKS cluster.",
		"nodes": "Specify which nodes winspect should send requests to using node names.",
		"pods":  "Specify which pods the capture should filter on. Supports up to two pod names. Automatically defines nodes to capture on.",
	}
	for _, name := range captureTypes {
		cmd := &cobra.Command{
			Use:   name,
			Short: captureHelp[name],
			Run:   getCapture,
		}

		captureCmd.AddCommand(cmd)
	}
}

func getCapture(cmd *cobra.Command, args []string) {
	hostMap := make(map[string][]string)

	// Revise nodes and pods arguments based on command name
	switch cmd.Name() {
	case "all":
		hostMap = comprise.CreateEmptyMap(targetNodes)
	case "nodes":
		if len(args) == 0 {
			log.Fatal("must pass node names when using 'winspect capture nodes ...'")
		}

		nodes := strings.Split(args[0], ",")
		targetNodes = client.ParseValidateNodes(nodes, nodeSet)
		hostMap = comprise.CreateEmptyMap(targetNodes)
	case "pods":
		if len(args) == 0 {
			log.Fatal("must pass pod names when using 'winspect capture pods ...'")
		}
		// populate node : pods map so that can be used in loop
		pods := strings.Split(args[0], ",")
		hostMap = client.ParseValidatePods(pods, podSet)
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

		cArgs := client.CaptureParams{
			Node:         ip,
			Pods:         hostMap[ip],
			Ips:          captureArgs.Ips,
			Protocols:    captureArgs.Protocols,
			Ports:        captureArgs.Ports,
			Macs:         captureArgs.Macs,
			Time:         captureArgs.Time,
			PacketType:   captureArgs.PacketType,
			CountersOnly: captureArgs.CountersOnly,
		}
		cArgs.ValidateCaptureParams()

		go client.RunCaptureStream(c, &cArgs, &wg)
	}

	wg.Wait()
}
