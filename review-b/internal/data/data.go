package data

import (
	"context"
	consul "github.com/go-kratos/kratos/contrib/registry/consul/v2"
	"github.com/go-kratos/kratos/v2/log"
	"github.com/go-kratos/kratos/v2/middleware/recovery"
	"github.com/go-kratos/kratos/v2/middleware/validate"
	"github.com/go-kratos/kratos/v2/registry"
	"github.com/go-kratos/kratos/v2/transport/grpc"
	"github.com/google/wire"
	"github.com/hashicorp/consul/api"
	v1 "review-b/api/review/v1"
	"review-b/internal/conf"
)

// ProviderSet is data providers.
var ProviderSet = wire.NewSet(NewData, NewBusinessRepo, NewReviewServiceClient, NewDiscovery)

// Data .
type Data struct {
	// TODO wrapped database client
	// 嵌入一个grpc Client端，去调用review-service服务
	rc  v1.ReviewClient
	log *log.Helper
}

// 直连服务，不通过注册中心
// func NewReviewServiceClient() v1.ReviewClient {
//	// 	"github.com/go-kratos/kratos/v2/transport/grpc"
//	conn, err := grpc.DialInsecure(context.Background(),
//		grpc.WithEndpoint("127.0.0.1:9000"),
//		grpc.WithMiddleware(recovery.Recovery(), validate.Validator()),
//	)
//	if err != nil {
//		panic(err)
//	}
//	return v1.NewReviewClient(conn)
//}

func NewDiscovery(conf *conf.Registry) registry.Discovery {
	client := api.DefaultConfig()
	// 使用配置文件中注册中心相关配置
	client.Address = conf.Consul.Address
	client.Scheme = conf.Consul.Scheme
	cli, err := api.NewClient(client)
	if err != nil {
		panic(err)
	}
	dis := consul.New(cli)
	return dis
}

// 注册中心，服务发现
func NewReviewServiceClient(d registry.Discovery) v1.ReviewClient {
	// 	"github.com/go-kratos/kratos/v2/transport/grpc"
	conn, err := grpc.DialInsecure(context.Background(),
		grpc.WithEndpoint("discovery:///review.service"),
		grpc.WithDiscovery(d),
		grpc.WithMiddleware(recovery.Recovery(), validate.Validator()),
	)
	if err != nil {
		panic(err)
	}
	return v1.NewReviewClient(conn)
}

// NewData .
func NewData(c *conf.Data, rc v1.ReviewClient, logger log.Logger) (*Data, func(), error) {
	cleanup := func() {
		log.NewHelper(logger).Info("closing the data resources")
	}
	return &Data{
		rc:  rc,
		log: log.NewHelper(logger),
	}, cleanup, nil
}
