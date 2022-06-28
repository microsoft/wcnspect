package cmd

import (
	"log"
	"sync"

	"github.com/microsoft/winspect/pkg/client"

	"github.com/spf13/cobra"
)

var counterCmd = &cobra.Command{
	Use:   "counter",
	Short: "The 'counter' command will retrieve packet counter tables from all windows nodes.",
	Long: `The 'counter' command will retrieve packet counter tables from all windows nodes. 
This command requires that a capture is being run on the requested nodes. For example:
'winspect counter --nodes {nodes} --include-hidden`,
	Run: getCounters,
}

func init() {
	var nodes []string
	var includeHidden bool

	rootCmd.AddCommand(counterCmd)

	counterCmd.PersistentFlags().StringSliceVarP(&nodes, "nodes", "n", []string{}, "Specify which nodes winspect should send requests to using node names. Runs on all windows nodes by default.")
	counterCmd.PersistentFlags().BoolVarP(&includeHidden, "include-hidden", "i", false, "Show counters from components that are hidden by default.")
}

func getCounters(cmd *cobra.Command, args []string) {
	nodes, err := cmd.Flags().GetStringSlice("nodes")
	if err != nil {
		log.Fatal(err)
	}

	includeHidden, err := cmd.Flags().GetBool("include-hidden")
	if err != nil {
		log.Fatal(err)
	}

	if len(nodes) != 0 {
		targetNodes = client.ParseValidateNodes(nodes, nodeSet)
	}

	var wg sync.WaitGroup
	for _, ip := range targetNodes {
		wg.Add(1)

		c, closeClient := client.CreateConnection(ip)
		defer closeClient()

		params := client.CounterParams{Node: ip, IncludeHidden: includeHidden}
		go client.PrintCounters(c, &params, &wg)
	}

	wg.Wait()
}
