package cmd

import (
	"sync"

	"github.com/microsoft/winspect/pkg/client"

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
	'winspect counter --nodes {nodes} --include-hidden`,
		Run: func(cmd *cobra.Command, args []string) {
			cc.printCounters()
		},
	}

	cmd.PersistentFlags().StringSliceVarP(&cc.nodes, "nodes", "n", []string{}, "Specify which nodes winspect should send requests to using node names. Runs on all windows nodes by default.")
	cmd.PersistentFlags().BoolVarP(&cc.includeHidden, "include-hidden", "i", false, "Show counters from components that are hidden by default.")

	cc.baseBuilderCmd = b.newBuilderCmd(cmd)

	return cc
}

func (cc *counterCmd) printCounters() {
	targetNodes := cc.getTargetNodes()

	if len(cc.nodes) > 0 {
		targetNodes = client.ParseValidateNodes(cc.nodes, cc.nodeSet)
	}

	var wg sync.WaitGroup
	for _, ip := range targetNodes {
		wg.Add(1)

		c, closeClient := client.CreateConnection(ip)
		defer closeClient()

		params := &client.CounterParams{Node: ip, IncludeHidden: cc.includeHidden}
		go client.PrintCounters(c, params, &wg)
	}

	wg.Wait()
}
