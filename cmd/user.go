package cmd

import (
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// userCmd represents the user command
var userCmd = &cobra.Command{
	Use:   "user",
	Short: "Actions to perform on a user",
	Run: func(cmd *cobra.Command, args []string) {
	},
}

func init() {
	rootCmd.AddCommand(userCmd)
	userCmd.AddCommand(createUserCmd)
	userCmd.AddCommand(deleteUserCmd)

	userCmd.PersistentFlags().StringP("user", "u", "", "the user to create")
	viper.BindPFlag("sysUser", userCmd.PersistentFlags().Lookup("user"))
}
