package cmd

import (
	"log"
	"os"
	"os/signal"
	"strings"
	"sync"
	"syscall"

	"github.com/microsoft/winspect/pkg/client"
	"github.com/microsoft/winspect/pkg/k8spi"
	pb "github.com/microsoft/winspect/rpc"

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
	cc.validateArgs()

	targetNodes := cc.getWinNodes()
	hostMap := make(map[string][]string)

	// Revise nodes and pods arguments based on command name
	switch subcmd {
	case "nodes":
		if len(endpoints) == 0 {
			log.Fatal("must pass node names when using 'winspect capture nodes ...'")
		}

		nodes := strings.Split(endpoints[0], ",")
		if err := client.ValidateNodes(nodes, cc.getWinNodeNames()); err != nil {
			log.Fatal(err)
		}

		targetNodes = cc.getNodes(nodes)
	case "pods":
		if len(endpoints) == 0 {
			log.Fatal("must pass pod names when using 'winspect capture pods ...'")
		}

		pods := strings.Split(endpoints[0], ",")
		if err := client.ValidatePods(pods, cc.getPodNames()); err != nil {
			log.Fatal(err)
		}

		hostMap = cc.getNodePodMap(pods)
		targetNodes = cc.getPodsNodes(pods)
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
	for _, node := range targetNodes {
		wg.Add(1)

		name, ip := node.GetName(), k8spi.RetrieveInternalIP(node)

		c, closeClient := client.CreateConnection(ip)
		defer closeClient()

		ctx := &client.ReqContext{
			Server: client.Node{
				Name: name,
				Ip:   ip,
			},
			Wg: &wg,
		}

		req := &pb.CaptureRequest{
			Duration: cc.time,
			Modifier: cc.getModifiers(hostMap[name]),
			Filter:   cc.getFilters(),
		}

		go client.RunCaptureStream(c, req, ctx)
	}

	wg.Wait()
}

func (cc *captureCmd) getFilters() *pb.Filters {
	return &pb.Filters{
		Ips:       cc.ips,
		Protocols: cc.protocols,
		Ports:     cc.ports,
		Macs:      cc.macs,
	}
}

func (cc *captureCmd) getModifiers(pods []string) *pb.Modifiers {
	return &pb.Modifiers{
		Pods:         pods,
		PacketType:   pb.PacketType(pb.PacketType_value[cc.packetType]),
		CountersOnly: cc.countersOnly,
	}
}

func (cc *captureCmd) validateArgs() {
	if err := client.ValidateTime(cc.time); err != nil {
		log.Fatal(err)
	}

	if err := client.ValidateIPAddrs(cc.ips); err != nil {
		log.Fatal(err)
	}

	if err := client.ValidateProtocols(cc.protocols); err != nil {
		log.Fatal(err)
	}

	if err := client.ValidatePorts(cc.ports); err != nil {
		log.Fatal(err)
	}

	if err := client.ValidateMACAddrs(cc.macs); err != nil {
		log.Fatal(err)
	}

	if err := client.ValidatePktType(cc.packetType); err != nil {
		log.Fatal(err)
	}
}
