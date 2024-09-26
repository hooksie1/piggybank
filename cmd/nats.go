package cmd

import (
	"github.com/CoverWhale/logr"
	"github.com/nats-io/jsm.go/natscontext"
	"github.com/nats-io/nats.go"
	"github.com/spf13/viper"
)

type natsOpts struct {
	name   string
	prefix string
}

func newNatsConnection(connOpts natsOpts) (*nats.Conn, error) {
	opts := []nats.Option{nats.Name(connOpts.name)}

	if connOpts.prefix != "" {
		opts = append(opts, nats.CustomInboxPrefix(connOpts.prefix))
	}

	if viper.GetString("credentials_file") == "" && viper.GetString("nats_user_jwt") == "" {
		logr.Debug("using NATS context")
		return natscontext.Connect("", opts...)
	}

	if viper.GetString("nats_user_jwt") != "" {
		logr.Debug("using env creds")
		opts = append(opts, nats.UserJWTAndSeed(viper.GetString("nats_user_jwt"), viper.GetString("nats_user_seed")))
	}
	if viper.GetString("credentials_file") != "" {
		opts = append(opts, nats.UserCredentials(viper.GetString("credentials_file")))
	}

	return nats.Connect(viper.GetString("nats_server"), opts...)
}
