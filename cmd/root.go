package cmd

import (
	"os"
	"strings"

	"github.com/CoverWhale/logr"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var cfgFile string
var cfg Config

var rootCmd = &cobra.Command{
	Use:   "piggybankctl",
	Short: "The app description",
	RunE:  start,
}
var replacer = strings.NewReplacer("-", "_")

type Config struct {
	Port int `mapstructure:"port"`
}

func Execute() {
	viper.SetDefault("service-name", "piggybank-local")
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}

func init() {
	cobra.OnInitialize(initConfig)
}

func initConfig() {

	viper.SetEnvPrefix("nex_hostservices")
	viper.AutomaticEnv()
	viper.SetEnvKeyReplacer(replacer)

	// If a config file is found, read it in.
	logger := logr.NewLogger()
	logger.Debug("initialized")
}
