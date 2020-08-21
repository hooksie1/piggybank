package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"

	homedir "github.com/mitchellh/go-homedir"
	"github.com/spf13/viper"
)

var cfgFile string
var url string

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "piggy",
	Short: "piggy is the cli utility for Piggy Bank",
	Long: `piggy can either start a Piggy Bank server or be
a cli agent to interact with an existing Piggy Bank server.`,
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func init() {
	cobra.OnInitialize(initConfig)

	// Here you will define your flags and configuration settings.
	// Cobra supports persistent flags, which, if defined here,
	// will be global for your application.

	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is $HOME/.piggybank.yml)")
	rootCmd.PersistentFlags().StringP("server", "s", "", "server loaded from config")
	rootCmd.PersistentFlags().BoolP("json", "j", false, "returns data in json format")
	viper.BindPFlag("server", rootCmd.PersistentFlags().Lookup("server"))
	viper.BindPFlag("jsonTrue", rootCmd.PersistentFlags().Lookup("json"))

}

// initConfig reads in config file and ENV variables if set.
func initConfig() {
	if cfgFile != "" {
		// Use config file from the flag.
		viper.SetConfigFile(cfgFile)
	} else {
		// Find home directory.
		home, err := homedir.Dir()
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}

		// Search config in home directory with name ".piggybank" (without extension).
		viper.AddConfigPath(home)
		viper.SetConfigName(".piggybank")
	}

	viper.AutomaticEnv() // read in environment variables that match

	// If a config file is found, read it in.

	if err := viper.ReadInConfig(); err != nil {
		fmt.Println("Could not find config file:", viper.ConfigFileUsed())
	}
}
