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
	Short: "The 'vfp-counter' command will retrieve packet counter tables from a specified windows pod's vNic's VFP.",
	Long: `The 'vfp-counter' command will retrieve packet counter tables from a specified windows pod's vNic's VFP. 
For example:
'winspect vfp-counter --pod {pod}'`,
	Run: getVFPCounters,
}

func init() {
	var pod string

	rootCmd.AddCommand(vfpcounterCmd)

	vfpcounterCmd.PersistentFlags().StringVarP(&pod, "pod", "p", "", "Specify which pods winspect should send requests to using pod names. This flag is required.")
	vfpcounterCmd.MarkPersistentFlagRequired("pod")
}

func getVFPCounters(cmd *cobra.Command, args []string) {
	pod, err := cmd.Flags().GetString("pod")
	if err != nil {
		log.Fatal(err)
	}

	hostMap := client.ParseValidatePods([]string{pod}, podSet)
	targetNodes := comprise.Keys(hostMap)

	var wg sync.WaitGroup
	for _, ip := range targetNodes {
		wg.Add(1)

		c, closeClient := client.CreateConnection(ip)
		defer closeClient()

		params := client.VFPCounterParams{Node: ip, Pod: hostMap[ip][0]}
		go client.PrintVFPCounters(c, &params, &wg)
	}

	wg.Wait()
}
