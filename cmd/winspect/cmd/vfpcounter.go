package cmd

import (
	"log"
	"sync"

	"github.com/microsoft/winspect/pkg/client"
	"github.com/microsoft/winspect/pkg/comprise"
	"github.com/spf13/cobra"
)

var vfpcounterCmd = &cobra.Command{
	Use:   "vfp-counter",
	Short: "The 'vfp-counter' command will retrieve packet counter tables from specified windows pods' vNic VFPs.",
	Long: `The 'vfp-counter' command will retrieve packet counter tables from specified windows pods' vNic VFPs. 
For example:
'winspect vfp-counter --pods {pods}`,
	Run: getVFPCounters,
}

func init() {
	var pods []string

	rootCmd.AddCommand(vfpcounterCmd)

	vfpcounterCmd.PersistentFlags().StringSliceVarP(&pods, "pods", "p", []string{}, "Specify which pods winspect should send requests to using pod names. This flag is required.")
	vfpcounterCmd.MarkPersistentFlagRequired("pods")
}

func getVFPCounters(cmd *cobra.Command, args []string) {
	pods, err := cmd.Flags().GetStringSlice("pods")
	if err != nil {
		log.Fatal(err)
	}

	hostMap := client.ParseValidatePods(pods, podSet)
	targetNodes = comprise.Keys(hostMap)

	var wg sync.WaitGroup
	for _, ip := range targetNodes {
		wg.Add(1)

		c, closeClient := client.CreateConnection(ip)
		defer closeClient()

		params := client.VFPCounterParams{Node: ip, Pods: hostMap[ip]}
		go client.PrintVFPCounters(c, &params, &wg)
	}

	wg.Wait()
}
