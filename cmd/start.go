package cmd

import (
	"fmt"

	cwnats "github.com/CoverWhale/coverwhale-go/transports/nats"
	"github.com/CoverWhale/logr"
	"github.com/hooksie1/piggybank/service"
	"github.com/invopop/jsonschema"
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
	//logger.SetOutput(cwnats.NewNatsLogger("prime.logs.piggybank", nc))

	svc, err := micro.AddService(nc, config)
	if err != nil {
		logr.Fatal(err)
	}

	dbGroup := svc.AddGroup("piggybank.database", micro.WithGroupQueueGroup("database"))
	dbGroup.AddEndpoint("initialize",
		service.AppHandler(logger, service.Initialize, appCtx),
		micro.WithEndpointMetadata(map[string]string{
			"description":     "initializes the database",
			"format":          "application/json",
			"request_schema":  "",
			"response_schema": schemaString(&service.ResponseMessage{}),
		}),
		micro.WithEndpointSubject("initialize"),
	)
	dbGroup.AddEndpoint("lock",
		service.AppHandler(logger, service.Lock, appCtx),
		micro.WithEndpointMetadata(map[string]string{
			"description":     "locks the database",
			"format":          "application/json",
			"request_schema":  "",
			"response_schema": schemaString(&service.ResponseMessage{}),
		}),
		micro.WithEndpointSubject("lock"),
	)
	dbGroup.AddEndpoint("unlock",
		service.AppHandler(logger, service.Unlock, appCtx),
		micro.WithEndpointMetadata(map[string]string{
			"description":     "unlocks the database",
			"format":          "application/json",
			"request_schema":  schemaString(&service.DatabaseKey{}),
			"response_schema": schemaString(&service.ResponseMessage{}),
		}),
	)

	appGroup := svc.AddGroup("piggybank.secrets", micro.WithGroupQueueGroup("app"))
	appGroup.AddEndpoint("GET",
		service.AppHandler(logger, service.GetRecord, appCtx),
		micro.WithEndpointMetadata(map[string]string{
			"description":     "Gets a secret",
			"format":          "application/json",
			"request_schema":  "",
			"response_schema": schemaString(&service.ResponseMessage{}),
		}),
		micro.WithEndpointSubject("GET.>"),
	)
	appGroup.AddEndpoint("POST",
		service.AppHandler(logger, service.AddRecord, appCtx),
		micro.WithEndpointMetadata(map[string]string{
			"description":     "Adds a secret",
			"format":          "application/json",
			"request_schema":  "",
			"response_schema": schemaString(&service.ResponseMessage{}),
		}),
		micro.WithEndpointSubject("POST.>"),
	)
	appGroup.AddEndpoint("DELETE",
		service.AppHandler(logger, service.DeleteRecord, appCtx),
		micro.WithEndpointMetadata(map[string]string{
			"description":     "Deletes a secret",
			"format":          "application/json",
			"request_schema":  "",
			"response_schema": schemaString(&service.ResponseMessage{}),
		}),
		micro.WithEndpointSubject("DELETE.>"),
	)

	// uncomment to enable config watching
	//go service.WatchForConfig(logger, js)

	logger.Infof("service %s %s started", svc.Info().Name, svc.Info().ID)

	health := func(ch chan<- string, s micro.Service) {
		a := <-nc.StatusChanged()
		ch <- fmt.Sprintf("%s %s", a.String(), nc.LastError())
	}

	return cwnats.HandleNotify(svc, health)

}

func schemaString(s any) string {
	schema := jsonschema.Reflect(s)
	data, err := schema.MarshalJSON()
	if err != nil {
		logr.Fatal(err)
	}

	return string(data)
}
