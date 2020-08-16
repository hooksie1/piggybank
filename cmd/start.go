package cmd

import (
	"github.com/hooksie1/piggybank/server"
	"github.com/spf13/cobra"
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
