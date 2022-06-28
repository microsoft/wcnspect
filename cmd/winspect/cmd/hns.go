package cmd

import (
	"log"
	"sync"

	"github.com/microsoft/winspect/pkg/client"

	"github.com/spf13/cobra"
)

var hnsCmd = &cobra.Command{
	Use:   "hns",
	Short: "The 'hns' command will retrieve hns logs on all windows nodes.",
	Long: `The 'hns' command will retrieve hns logs on all windows nodes. For example:
'winspect hns all --nodes {nodes} --json`,
}

func init() {
	var nodes []string
	var verbose bool

	rootCmd.AddCommand(hnsCmd)
	hnsCmd.PersistentFlags().StringSliceVarP(&nodes, "nodes", "n", []string{}, "Specify which nodes winspect should send requests to using node names. Runs on all windows nodes by default.")
	hnsCmd.PersistentFlags().BoolVarP(&verbose, "json", "d", false, "Detailed option for logs.")

	logTypes := []string{"all", "endpoints", "loadbalancers", "namespaces", "networks"}
	logHelp := map[string]string{
		"all":           "Retrieve all hns logs on each node.",
		"endpoints":     "Retrieve logs for endpoints on each node.",
		"loadbalancers": "Retrieve logs for loadbalancers on each node.",
		"namespaces":    "Retrieve logs for namespaces on each node.",
		"networks":      "Retrieve logs for networks on each node.",
	}
	for _, name := range logTypes {
		cmd := &cobra.Command{
			Use:   name,
			Short: logHelp[name],
			Run:   getLogs,
		}

		hnsCmd.AddCommand(cmd)
	}
}

func getLogs(cmd *cobra.Command, args []string) {
	nodes, err := cmd.Flags().GetStringSlice("nodes")
	if err != nil {
		log.Fatal(err)
	}

	verbose, err := cmd.Flags().GetBool("json")
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

		params := client.HCNParams{Cmd: cmd.Name(), Node: ip, Verbose: verbose}
		go client.PrintHCNLogs(c, &params, &wg)
	}

	wg.Wait()
}
