package service

import (
	"github.com/CoverWhale/logr"
	"github.com/nats-io/nats.go/micro"
)

func DBGroup(svc micro.Service, logger *logr.Logger, appCtx AppContext) {
	dbGroup := svc.AddGroup(databaseSubject, micro.WithGroupQueueGroup("database"))
	dbGroup.AddEndpoint("initialize",
		AppHandler(logger, Initialize, appCtx),
		micro.WithEndpointMetadata(map[string]string{
			"description": "initializes the database",
			"format":      "application/json",
		}),
		micro.WithEndpointSubject(databaseInitSubject),
	)
	dbGroup.AddEndpoint("status",
		AppHandler(logger, SecretHandler(Status), appCtx),
		micro.WithEndpointMetadata(map[string]string{
			"description": "returns the status of the database",
			"format":      "application/json",
		}),
		micro.WithEndpointSubject(databaseStatusSubject),
	)
	dbGroup.AddEndpoint("lock",
		AppHandler(logger, Lock, appCtx),
		micro.WithEndpointMetadata(map[string]string{
			"description": "locks the database",
			"format":      "application/json",
		}),
		micro.WithEndpointSubject(databaseLockSubject),
	)
	dbGroup.AddEndpoint("unlock",
		AppHandler(logger, Unlock, appCtx),
		micro.WithEndpointMetadata(map[string]string{
			"description": "unlocks the database",
			"format":      "application/json",
		}),
		micro.WithEndpointSubject(databaseUnlockSubject),
	)
	dbGroup.AddEndpoint("rotate",
		AppHandler(logger, RotateKey, appCtx),
		micro.WithEndpointMetadata(map[string]string{
			"description": "rotates the database encryption key",
			"format":      "application/json",
		}),
		micro.WithEndpointSubject(databaseRotateSubject),
	)
}

func AppGroup(svc micro.Service, logger *logr.Logger, appCtx AppContext) {
	appGroup := svc.AddGroup(secretSubject, micro.WithGroupQueueGroup("app"))
	appGroup.AddEndpoint("GET",
		AppHandler(logger, SecretHandler(GetRecord), appCtx),
		micro.WithEndpointMetadata(map[string]string{
			"description": "Gets a secret",
			"format":      "application/json",
		}),
		micro.WithEndpointSubject("GET.>"),
	)
	appGroup.AddEndpoint("POST",
		AppHandler(logger, SecretHandler(AddRecord), appCtx),
		micro.WithEndpointMetadata(map[string]string{
			"description": "Adds a secret",
			"format":      "application/json",
		}),
		micro.WithEndpointSubject("POST.>"),
	)
	appGroup.AddEndpoint("DELETE",
		AppHandler(logger, SecretHandler(DeleteRecord), appCtx),
		micro.WithEndpointMetadata(map[string]string{
			"description": "Deletes a secret",
			"format":      "application/json",
		}),
		micro.WithEndpointSubject("DELETE.>"),
	)
}
