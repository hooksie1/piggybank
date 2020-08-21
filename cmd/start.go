package cmd

import (
	"github.com/spf13/cobra"
	"gitlab.com/hooksie1/piggybank/server"
)

// startCmd represents the start command
var startCmd = &cobra.Command{
	Use:   "start",
	Short: "Start the piggy bank webserver",
	Run: func(cmd *cobra.Command, args []string) {
		server.Serve()
	},
}

func init() {
	rootCmd.AddCommand(startCmd)
}
