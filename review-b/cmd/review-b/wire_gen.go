// Code generated by Wire. DO NOT EDIT.

//go:generate go run -mod=mod github.com/google/wire/cmd/wire
//go:build !wireinject
// +build !wireinject

package main

import (
	"github.com/go-kratos/kratos/v2"
	"github.com/go-kratos/kratos/v2/log"
	"review-b/internal/biz"
	"review-b/internal/conf"
	"review-b/internal/data"
	"review-b/internal/server"
	"review-b/internal/service"
)

import (
	_ "go.uber.org/automaxprocs"
)

// Injectors from wire.go:

// wireApp init kratos application.
func wireApp(confServer *conf.Server, registry *conf.Registry, confData *conf.Data, logger log.Logger) (*kratos.App, func(), error) {
	discovery := data.NewDiscovery(registry)
	reviewClient := data.NewReviewServiceClient(discovery)
	dataData, cleanup, err := data.NewData(confData, reviewClient, logger)
	if err != nil {
		return nil, nil, err
	}
	businessRepo := data.NewBusinessRepo(dataData, logger)
	businessUseCase := biz.NewBusinessUseCase(businessRepo, logger)
	businessService := service.NewBusinessService(businessUseCase)
	grpcServer := server.NewGRPCServer(confServer, businessService, logger)
	httpServer := server.NewHTTPServer(confServer, businessService, logger)
	app := newApp(logger, grpcServer, httpServer)
	return app, func() {
		cleanup()
	}, nil
}
