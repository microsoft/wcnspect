// Copyright (c) Microsoft Corporation.
// Licensed under the MIT license.

package cmd

import (
	"log"
	"sync"

	"github.com/microsoft/wcnspect/pkg/client"
	"github.com/microsoft/wcnspect/pkg/k8spi"
	pb "github.com/microsoft/wcnspect/rpc"

	"github.com/spf13/cobra"
)

type counterCmd struct {
	nodes         []string
	includeHidden bool

	*baseBuilderCmd
}

func (b *commandsBuilder) newCounterCmd() *counterCmd {
	cc := &counterCmd{}

	cmd := &cobra.Command{
		Use:   "counter",
		Short: "The 'counter' command will retrieve packet counter tables from all windows nodes.",
		Long: `The 'counter' command will retrieve packet counter tables from all windows nodes. 
	This command requires that a capture is being run on the requested nodes. For example:
	'wcnspect counter --nodes {nodes} --include-hidden`,
		Run: func(cmd *cobra.Command, args []string) {
			cc.printCounters()
		},
	}

	cmd.PersistentFlags().StringSliceVarP(&cc.nodes, "nodes", "n", []string{}, "Specify which nodes wcnspect should send requests to using node names. Runs on all windows nodes by default.")
	cmd.PersistentFlags().BoolVarP(&cc.includeHidden, "include-hidden", "i", false, "Show counters from components that are hidden by default.")

	cc.baseBuilderCmd = b.newBuilderCmd(cmd)

	return cc
}

func (cc *counterCmd) printCounters() {
	targetNodes := cc.getWinNodes()

	if len(cc.nodes) > 0 {
		if err := client.ValidateNodes(cc.nodes, cc.getWinNodeNames()); err != nil {
			log.Fatal(err)
		}

		targetNodes = cc.getNodes(cc.nodes)
	}

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

		req := &pb.CountersRequest{
			IncludeHidden: cc.includeHidden,
		}

		go client.PrintCounters(c, req, ctx)
	}

	wg.Wait()
}
