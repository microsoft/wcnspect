// Copyright (c) Microsoft Corporation.
// Licensed under the MIT license.

package cmd

import (
	"log"
	"sync"

	"github.com/microsoft/winspect/pkg/client"
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
	pods := []string{cc.pod}

	if err := client.ValidatePods(pods, cc.getPodNames()); err != nil {
		log.Fatal(err)
	}

	hostMap := cc.getNodePodMap(pods)
	targetNodes := cc.getPodsNodes(pods)

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

		req := &pb.VFPCountersRequest{
			Pod:     hostMap[name][0],
			Verbose: cc.verbose,
		}

		go client.PrintVFPCounters(c, req, ctx)
	}

	wg.Wait()
}
