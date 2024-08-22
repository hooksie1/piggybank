package cmd

import (
	"fmt"

	cwnats "github.com/CoverWhale/coverwhale-go/transports/nats"
	"github.com/CoverWhale/logr"
	"github.com/hooksie1/piggybank/service"
	"github.com/nats-io/nats.go"
	"github.com/nats-io/nats.go/micro"
	"github.com/spf13/cobra"
)

var startCmd = &cobra.Command{
	Use:          "start",
	Short:        "starts the service",
	RunE:         start,
	SilenceUsage: true,
}

func init() {
	// attach start subcommand to service subcommand
	serviceCmd.AddCommand(startCmd)
}

func start(cmd *cobra.Command, args []string) error {
	logger := logr.NewLogger()

	config := micro.Config{
		Name:        "piggybank",
		Version:     "0.0.1",
		Description: "Secrets storage for NATS",
	}

	nc, err := newNatsConnection("piggybank-server")
	if err != nil {
		return err
	}
	defer nc.Close()

	js, err := nc.JetStream()
	if err != nil {
		return err
	}

	kv, err := js.KeyValue(service.Bucket)
	if err != nil {
		return err
	}

	appCtx := service.AppContext{
		KV: kv,
	}

	// uncomment for config watching
	//js, err := nc.JetStream()
	//if err != nil {
	//    return err
	//}

	// uncomment to enable logging over NATS
	//logger.SetOutput(cwnats.NewNatsLogger("logs.piggybank", nc))

	svc, err := micro.AddService(nc, config)
	if err != nil {
		logr.Fatal(err)
	}

	service.DBGroup(svc, logger, appCtx)
	service.AppGroup(svc, logger, appCtx)

	// uncomment to enable config watching
	//go service.WatchForConfig(logger, js)

	logger.Infof("service %s %s started", svc.Info().Name, svc.Info().ID)

	health := func(ch chan<- string, s micro.Service) {
		a := <-nc.StatusChanged()
		if a == nats.CLOSED {
			ch <- fmt.Sprintf("%s last error: %v", a.String(), nc.LastError())
		}
	}

	return cwnats.HandleNotify(svc, health)
}
