package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

var Version = "dev"

// versionCmd represents the version command
var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Prints the version",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println(Version)
	},
}

func init() {
	rootCmd.AddCommand(versionCmd)
}
