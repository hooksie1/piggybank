package cmd

import (
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

//Flags are defined here. Because of the way Viper binds values, if the same flag name is called
// with viper.BindPFlag multiple times during init() the value will be overwritten. For example if
// two subcommands each have a flag called name but they each have their own default values,
// viper can overwrite any value passed in for one subcommand with the default value of the other subcommand.
// The answer here is to not use init() and instead use something like PersistentPreRun to bind the
// viper values. Using init for the cobra flags is ok, they are only in here to limit duplication of names.

// bindNatsFlags binds nats flag values to viper
func bindNatsFlags(cmd *cobra.Command) {
	viper.BindPFlag("nats_urls", cmd.Flags().Lookup("nats-urls"))
	viper.BindPFlag("nats_seed", cmd.Flags().Lookup("nats-seed"))
	viper.BindPFlag("nats_jwt", cmd.Flags().Lookup("nats-jwt"))
	viper.BindPFlag("nats_secret", cmd.Flags().Lookup("nats-secret"))
	viper.BindPFlag("credentials_file", cmd.Flags().Lookup("credentials-file"))
	viper.BindPFlag("use_traffic_shaping", cmd.Flags().Lookup("use-traffic-shaping"))
}

// natsFlags adds the nats flags to the passed in cobra command
func natsFlags(cmd *cobra.Command) {
	cmd.PersistentFlags().String("nats-jwt", "", "NATS JWT as a string")
	cmd.PersistentFlags().String("nats-seed", "", "NATS seed as a string")
	cmd.PersistentFlags().String("credentials-file", "", "Path to NATS user credentials file")
	cmd.PersistentFlags().String("nats-urls", "nats://localhost:4222", "NATS URLs")
	cmd.PersistentFlags().Bool("use-traffic-shaping", false, "Local development connection")
}

func bindClientFlags(cmd *cobra.Command) {
	viper.BindPFlag("inbox_prefix", cmd.Flags().Lookup("inbox-prefix"))
}

func clientFlags(cmd *cobra.Command) {
	cmd.PersistentFlags().String("inbox-prefix", "PIGGYBANK.ADMIN", "subject prefix for replies")
}
