package cmd

import (
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

type AppUser struct {
	Application string `json:"application"`
	Username    string `json:"username"`
	Password    string `json:"password"`
}

// credCmd represents the cred command
var credCmd = &cobra.Command{
	Use:   "cred",
	Short: "Actions to perform on a credential",
	Run: func(cmd *cobra.Command, args []string) {
	},
}

func init() {
	rootCmd.AddCommand(credCmd)
	credCmd.AddCommand(createPassCmd)
	credCmd.AddCommand(lookupCmd)
	credCmd.AddCommand(deletePassCmd)

	credCmd.PersistentFlags().StringVarP(&application, "app", "a", "", "the application for the credential")
	credCmd.PersistentFlags().StringVarP(&user, "user", "u", "", "the application username")
	viper.BindPFlag("app", credCmd.PersistentFlags().Lookup("app"))
	viper.BindPFlag("user", credCmd.PersistentFlags().Lookup("user"))
}
