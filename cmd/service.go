package cmd

import (
	"github.com/spf13/cobra"
)

var serviceCmd = &cobra.Command{
	Use:   "service",
	Short: "subcommand to control the service",
	// PersistentPostRun is used here because this is just a subcommand with no run function
	PersistentPreRun: bindServiceCmdFlags,
}

func init() {
	rootCmd.AddCommand(serviceCmd)
	natsFlags(serviceCmd)
}

func bindServiceCmdFlags(cmd *cobra.Command, args []string) {
	bindNatsFlags(cmd)
}
