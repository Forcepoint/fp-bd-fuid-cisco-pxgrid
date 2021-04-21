// pxgrid command defines that the integration is going to be with Cisco ISE pxGrid service

package cmd

import (
	"github.com/spf13/cobra"
	"os"
)

// pxgridCmd represents the pxgrid command
var pxgridCmd = &cobra.Command{
	Use:   "pxgrid",
	Short: "Cisco PxGrid service",
	Long:  `Cisco PxGrid service. sub-commands {create-client, consumer}`,
	Run: func(cmd *cobra.Command, args []string) {
		_ = cmd.Help()
		os.Exit(0)
	},
}

func init() {
	rootCmd.AddCommand(pxgridCmd)
}
