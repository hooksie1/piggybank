package cmd

import (
	"os"

	"github.com/CoverWhale/logr"
	"github.com/nats-io/jsm.go/natscontext"
	"github.com/nats-io/nats.go"
	"github.com/spf13/viper"
)

func newNatsConnection(name string) (*nats.Conn, error) {
	opts := []nats.Option{nats.Name(name), nats.CustomInboxPrefix(viper.GetString("inbox_prefix"))}

	_, ok := os.LookupEnv("USER")

	if viper.GetString("credentials_file") == "" && viper.GetString("nats_jwt") == "" && ok {
		logr.Debug("using NATS context")
		return natscontext.Connect("", opts...)
	}

	if viper.GetString("nats_jwt") != "" {
		opts = append(opts, nats.UserJWTAndSeed(viper.GetString("nats_jwt"), viper.GetString("nats_seed")))
	}
	if viper.GetString("credentials_file") != "" {
		opts = append(opts, nats.UserCredentials(viper.GetString("credentials_file")))
	}

	return nats.Connect(viper.GetString("nats_urls"), opts...)
}
