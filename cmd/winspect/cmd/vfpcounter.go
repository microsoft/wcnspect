package cmd

import (
	"sync"

	"github.com/microsoft/winspect/pkg/client"
	"github.com/microsoft/winspect/pkg/comprise"
	"github.com/microsoft/winspect/pkg/k8spi"
	pb "github.com/microsoft/winspect/rpc"

	"github.com/spf13/cobra"
)

type vfpCounterCmd struct {
	pod     string
	verbose bool

	*baseBuilderCmd
}

func (b *commandsBuilder) newVfpCounterCmd() *vfpCounterCmd {
	cc := &vfpCounterCmd{}

	cmd := &cobra.Command{
		Use:   "vfp-counter",
		Short: "The 'vfp-counter' command will retrieve packet counter tables from a specified windows pod's port VFP.",
		Long: `The 'vfp-counter' command will retrieve packet counter tables from a specified windows pod's port VFP. 
	For example:
	'winspect vfp-counter --pod {pod}'`,
		Run: func(cmd *cobra.Command, args []string) {
			cc.printVFPCounters()
		},
	}

	cmd.PersistentFlags().StringVarP(&cc.pod, "pod", "p", "", "Specify which pod winspect should send requests to using pod name. This flag is required.")
	cmd.PersistentFlags().BoolVarP(&cc.verbose, "detailed", "d", false, "Option to output Host vNic and External Adapter Port counters.")
	cmd.MarkPersistentFlagRequired("pod")

	cc.baseBuilderCmd = b.newBuilderCmd(cmd)

	return cc
}

func (cc *vfpCounterCmd) printVFPCounters() {
	hostMap := client.ParseValidatePods([]string{cc.pod}, cc.podSet)
	targetNodes := comprise.Keys(hostMap)

	var wg sync.WaitGroup
	for _, ip := range targetNodes {
		wg.Add(1)

		c, closeClient := client.CreateConnection(ip)
		defer closeClient()

		ctx := &client.ReqContext{
			Server: client.Node{
				Name: k8spi.GetNodesIpToName(cc.nodeSet)[ip], //FIXME: move parsing to commands.go
				Ip:   ip,
			},
			Wg: &wg,
		}

		req := &pb.VFPCountersRequest{
			Pod:     hostMap[ip][0],
			Verbose: cc.verbose,
		}

		go client.PrintVFPCounters(c, req, ctx)
	}

	wg.Wait()
}
