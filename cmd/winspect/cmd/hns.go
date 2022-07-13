package cmd

import (
	"sync"

	"github.com/microsoft/winspect/pkg/client"

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
	'winspect hns all --nodes {nodes} --json`,
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

	cmd.PersistentFlags().StringSliceVarP(&cc.nodes, "nodes", "n", []string{}, "Specify which nodes winspect should send requests to using node names. Runs on all windows nodes by default.")
	cmd.PersistentFlags().BoolVarP(&cc.verbose, "json", "d", false, "Detailed option for logs.")

	cc.baseBuilderCmd = b.newBuilderCmd(cmd)

	return cc
}

func (cc *hnsCmd) printLogs(subcmd string) {
	targetNodes := cc.getTargetNodes()

	if len(cc.nodes) != 0 {
		targetNodes = client.ParseValidateNodes(cc.nodes, cc.nodeSet)
	}

	var wg sync.WaitGroup
	for _, ip := range targetNodes {
		wg.Add(1)

		c, closeClient := client.CreateConnection(ip)
		defer closeClient()

		params := &client.HCNParams{Cmd: subcmd, Node: ip, Verbose: cc.verbose}
		go client.PrintHCNLogs(c, params, &wg)
	}

	wg.Wait()
}
