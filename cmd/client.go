package cmd

import (
	"github.com/spf13/cobra"
)

var clientCmd = &cobra.Command{
	Use:              "client",
	Short:            "Client interactions with the service",
	PersistentPreRun: bindClientCmdFlags,
}

func init() {
	rootCmd.AddCommand(clientCmd)
	natsFlags(clientCmd)
}

func bindClientCmdFlags(cmd *cobra.Command, args []string) {
	bindNatsFlags(cmd)
}
