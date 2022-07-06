package cmd

import (
	"fmt"

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
	fmt.Println("vfp counters.") //TODO: remove
}
