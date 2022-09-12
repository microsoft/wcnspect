// Copyright (c) Microsoft Corporation.
// Licensed under the MIT license.

package cmd

import (
	"log"
	"sync"

	"github.com/microsoft/wcnspect/pkg/client"
	"github.com/microsoft/wcnspect/pkg/k8sapi"
	pb "github.com/microsoft/wcnspect/rpc"

	"github.com/spf13/cobra"
)

type hnsCmd struct {
	nodes   []string
	verbose bool

	*baseBuilderCmd
}

func (b *commandsBuilder) newHnsCmd() *hnsCmd {
	cc := &hnsCmd{}

	cmd := &cobra.Command{
		Use:   "hns",
		Short: "The 'hns' command will retrieve hns logs on all windows nodes.",
		Long: `The 'hns' command will retrieve hns logs on all windows nodes. For example:
	'wcnspect hns all --nodes {nodes} --json`,
	}

	logTypes := []string{"all", "endpoints", "loadbalancers", "namespaces", "networks"}
	logHelp := map[string]string{
		"all":           "Retrieve all hns logs on each node.",
		"endpoints":     "Retrieve logs for endpoints on each node.",
		"loadbalancers": "Retrieve logs for loadbalancers on each node.",
		"namespaces":    "Retrieve logs for namespaces on each node.",
		"networks":      "Retrieve logs for networks on each node.",
	}
	for _, name := range logTypes {
		subcmd := &cobra.Command{
			Use:   name,
			Short: logHelp[name],
			Run: func(cmd *cobra.Command, args []string) {
				cc.printLogs(cmd.Name())
			},
		}

		cmd.AddCommand(subcmd)
	}

	cmd.PersistentFlags().StringSliceVarP(&cc.nodes, "nodes", "n", []string{}, "Specify which nodes wcnspect should send requests to using node names. Runs on all windows nodes by default.")
	cmd.PersistentFlags().BoolVarP(&cc.verbose, "json", "d", false, "Detailed option for logs.")

	cc.baseBuilderCmd = b.newBuilderCmd(cmd)

	return cc
}

func (cc *hnsCmd) printLogs(subcmd string) {
	targetNodes := cc.getWinNodes()

	if len(cc.nodes) != 0 {
		if err := client.ValidateNodes(cc.nodes, cc.getWinNodeNames()); err != nil {
			log.Fatal(err)
		}

		targetNodes = cc.getNodes(cc.nodes)
	}

	var wg sync.WaitGroup
	for _, node := range targetNodes {
		wg.Add(1)

		name, ip := node.GetName(), k8sapi.RetrieveInternalIP(node)

		c, closeClient := client.CreateConnection(ip)
		defer closeClient()

		ctx := &client.ReqContext{
			Server: client.Node{
				Name: name,
				Ip:   ip,
			},
			Wg: &wg,
		}

		req := &pb.HCNRequest{
			Hcntype: pb.HCNType(pb.HCNType_value[subcmd]),
			Verbose: cc.verbose,
		}

		go client.PrintHCNLogs(c, req, ctx)
	}

	wg.Wait()
}
